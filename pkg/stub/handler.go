// Copyright 2018 The rethinkdb-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stub

import (
	"fmt"
	"os"
	"reflect"

	"github.com/coreos/operator-sdk/pkg/sdk/action"
	"github.com/coreos/operator-sdk/pkg/sdk/handler"
	"github.com/coreos/operator-sdk/pkg/sdk/query"
	"github.com/coreos/operator-sdk/pkg/sdk/types"
	v1alpha1 "github.com/jmckind/rethinkdb-operator/pkg/apis/operator/v1alpha1"
	"github.com/jmckind/rethinkdb-operator/pkg/util/k8sutil"
	"github.com/sirupsen/logrus"
	apps_v1beta2 "k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
)

const (
	defaultConfig = `bind=all
directory=/var/lib/rethinkdb/default
`
)

func NewRethinkDBHandler() handler.Handler {
	return &RethinkDBHandler{
		namespace: os.Getenv("MY_POD_NAMESPACE"),
		kubecli:	 k8sutil.MustNewKubeClient(),
	}
}

type RethinkDBHandler struct {
	namespace string
	kubecli		kubernetes.Interface
}

func (h *RethinkDBHandler) Handle(ctx types.Context, event types.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.RethinkDB:
		err := h.HandleRethinkDB(o)
		if err != nil {
			return fmt.Errorf("failed to handle rethinkdb: %v", err)
		}
	}
	return nil
}

func (h *RethinkDBHandler) HandleRethinkDB(r *v1alpha1.RethinkDB) error {
	logrus.Infof("handling rethinkdb: %v", r)

	labels := map[string]string{
		"app":  "rethinkdb",
		"cluster": r.Name,
	}

	r.SetDefaults()

	err := h.CreateOrUpdateClusterService(r, labels)
	if err != nil {
		return fmt.Errorf("failed to create or update cluster service: %v", err)
	}

	err = h.CreateOrUpdateDriverService(r, labels)
	if err != nil {
		return fmt.Errorf("failed to create or update driver service: %v", err)
	}

	err = h.CreateOrUpdateStatefulSet(r, labels)
	if err != nil {
		return fmt.Errorf("failed to create or update statefulset: %v", err)
	}

	err = h.UpdateStatus(r)
	if err != nil {
		return fmt.Errorf("failed to update rethinkdb status: %v", err)
	}

	return nil
}

func (h *RethinkDBHandler) CreateOrUpdateStatefulSet(r *v1alpha1.RethinkDB, labels map[string]string) error {
	name := r.Name
	replicas := r.Spec.Size
	cluster := name + "-cluster"
	terminationSeconds := int64(5)

	ss := &apps_v1beta2.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1beta2",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: r.ObjectMeta.Namespace,
			Labels: labels,
		},
		Spec: apps_v1beta2.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			ServiceName: cluster,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Containers: h.AddContainers(r),
					InitContainers: h.AddInitContainers(r),
					TerminationGracePeriodSeconds: &terminationSeconds,
					Volumes: h.AddVolumes(r),
				},
			},
			VolumeClaimTemplates: h.AddPVCs(r),
		},
	}

	err := action.Create(ss)
	if err != nil && !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create statefulset: %v", err)
	}

	err = query.Get(ss)
	if err != nil {
		return fmt.Errorf("failed to get statefulset: %v", err)
	}

	if *ss.Spec.Replicas != replicas {
		logrus.Infof("updating statefulset: %v", name)
		ss.Spec.Replicas = &replicas
		err = action.Update(ss)
		if err != nil {
			return fmt.Errorf("failed to update statefulset: %v", err)
		}
	}

	return nil
}

func (h *RethinkDBHandler) AddContainers(r *v1alpha1.RethinkDB) []v1.Container {
	return []v1.Container{{
		Command: []string{
			"/usr/bin/rethinkdb",
			"--no-update-check",
			"--config-file",
			"/etc/rethinkdb/rethinkdb.conf",
		},
		Image: fmt.Sprintf("%s:%s", r.Spec.BaseImage, r.Spec.Version),
		Name: "rethinkdb",
		Ports: []v1.ContainerPort{{
			ContainerPort: 8080,
			Name:          "http",
		},
		{
			ContainerPort: 28015,
			Name:          "driver",
		},
		{
			ContainerPort: 29015,
			Name:          "cluster",
		}},
		Resources: r.Spec.Pod.Resources,
		Stdin: true,
		TTY: true,
		VolumeMounts: []v1.VolumeMount{{
			Name:      "rethinkdb-data",
			MountPath: "/var/lib/rethinkdb/default",
		},
		{
			Name:      "rethinkdb-etc",
			MountPath: "/etc/rethinkdb",
		}},
	}}
}

func (h *RethinkDBHandler) AddInitContainers(r *v1alpha1.RethinkDB) []v1.Container {
	name := r.Name
	cluster := name + "-cluster"

	return []v1.Container{{
		Command: []string{
			"/bin/sh",
			"-c",
			fmt.Sprintf("echo '%s' > /etc/rethinkdb/rethinkdb.conf; if nslookup %s; then echo join=%s-0.%s:29015 >> /etc/rethinkdb/rethinkdb.conf; fi;", defaultConfig, cluster, name, cluster),
		},
		Image: "busybox:latest",
		Name: "cluster-init",
		VolumeMounts: []v1.VolumeMount{{
			Name:      "rethinkdb-etc",
			MountPath: "/etc/rethinkdb",
		}},
	}}
}

func (h *RethinkDBHandler) AddVolumes(r *v1alpha1.RethinkDB) []v1.Volume {
	var volumes []v1.Volume

	volumes = append(volumes, h.AddEmptyDirVolume("rethinkdb-etc"))

	if !r.IsPVEnabled() {
		volumes = append(volumes, h.AddEmptyDirVolume("rethinkdb-data"))
	}

	return volumes
}

func (h *RethinkDBHandler) AddEmptyDirVolume(name string) v1.Volume {
	return v1.Volume{
		Name: name,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}
}

func (h *RethinkDBHandler) AddPVCs(r *v1alpha1.RethinkDB) []v1.PersistentVolumeClaim {
	var pvcs []v1.PersistentVolumeClaim

	if r.IsPVEnabled() {
		pvcs = append(pvcs, v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rethinkdb-data",
				Namespace: r.ObjectMeta.Namespace,
				Labels: r.ObjectMeta.Labels,
			},
			Spec: *r.Spec.Pod.PersistentVolumeClaimSpec,
		})
	}

	return pvcs
}

func (h *RethinkDBHandler) CreateOrUpdateClusterService(r *v1alpha1.RethinkDB, labels map[string]string) error {
	name := r.Name + "-cluster"
	svc := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: r.ObjectMeta.Namespace,
			Labels: labels,
		},
	}

	err := query.Get(svc)
	if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get cluster service: %v", err)
	}

	if apierrors.IsNotFound(err) {
		logrus.Infof("creating cluster service: %v", name)
		svc.Spec = v1.ServiceSpec{
			ClusterIP: "None",
			Selector: labels,
			Ports: []v1.ServicePort{{
				Port: 29015,
				Name: "cluster",
			}},
		}

		err = action.Create(svc)
		if err != nil {
			return fmt.Errorf("failed to create cluster service: %v", err)
		}
	}

	return nil
}

func (h *RethinkDBHandler) CreateOrUpdateDriverService(r *v1alpha1.RethinkDB, labels map[string]string) error {
	svc := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name,
			Namespace: r.ObjectMeta.Namespace,
			Labels: labels,
		},
	}

	err := query.Get(svc)
	if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get driver service: %v", err)
	}

	if apierrors.IsNotFound(err) {
		logrus.Infof("creating driver service: %v", r.Name)
		svc.Spec = v1.ServiceSpec{
			Selector: labels,
			SessionAffinity: "ClientIP",
			Type: "NodePort",
			Ports: []v1.ServicePort{{
				Port: 8080,
				Name: "http",
			},
			{
				Port: 28015,
				Name: "driver",
			}},
		}

		err = action.Create(svc)
		if err != nil {
			return fmt.Errorf("failed to create driver service: %v", err)
		}
	}

	return nil
}

func (h *RethinkDBHandler) UpdateStatus(r *v1alpha1.RethinkDB) error {
	var podNames []string
	podList := &v1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
	}

	labelSelector := apilabels.SelectorFromSet(r.ObjectMeta.Labels).String()
	listOps := &metav1.ListOptions{LabelSelector: labelSelector}

	err := query.List(r.ObjectMeta.Namespace, podList, query.WithListOptions(listOps))
	if err != nil {
		return fmt.Errorf("failed to list pods: %v", err)
	}

	for _, pod := range podList.Items {
		podNames = append(podNames, pod.Name)
	}

	if !reflect.DeepEqual(podNames, r.Status.Pods) {
		r.Status.Pods = podNames
		err := action.Update(r)
		if err != nil {
			return fmt.Errorf("failed to update rethinkdb status: %v", err)
		}
	}

	return nil
}
