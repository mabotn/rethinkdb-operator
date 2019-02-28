// Copyright 2018 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package generate

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	genutil "github.com/operator-framework/operator-sdk/commands/operator-sdk/cmd/generate/internal"
	"github.com/operator-framework/operator-sdk/internal/util/projutil"
	"github.com/operator-framework/operator-sdk/pkg/scaffold"
	"github.com/operator-framework/operator-sdk/pkg/scaffold/input"

	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var headerFile string

func NewGenerateOpenAPICmd() *cobra.Command {
	openAPICmd := &cobra.Command{
		Use:   "openapi",
		Short: "Generates OpenAPI specs for API's",
		Long: `generate openapi generates OpenAPI validation specs in Go from tagged types
in all pkg/apis/<group>/<version> directories. Go code is generated under
pkg/apis/<group>/<version>/zz_generated.openapi.go. CRD's are generated, or
updated if they exist for a particular group + version + kind, under
deploy/crds/<group>_<version>_<kind>_crd.yaml; OpenAPI V3 validation YAML
is generated as a 'validation' object.

Example:
	$ operator-sdk generate openapi
	$ tree pkg/apis
	pkg/apis/
	└── app
		└── v1alpha1
			├── zz_generated.openapi.go
	$ tree deploy/crds
	├── deploy/crds/app_v1alpha1_appservice_cr.yaml
	├── deploy/crds/app_v1alpha1_appservice_crd.yaml
`,
		RunE: openAPIFunc,
	}

	openAPICmd.Flags().StringVar(&headerFile, "header-file", "", "Path to file containing headers for generated files.")

	return openAPICmd
}

func openAPIFunc(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("command %s doesn't accept any arguments", cmd.CommandPath())
	}

	return OpenAPIGen()
}

// OpenAPIGen generates OpenAPI validation specs for all CRD's in dirs.
func OpenAPIGen() error {
	projutil.MustInProjectRoot()

	absProjectPath := projutil.MustGetwd()
	repoPkg := projutil.CheckAndGetProjectGoPkg()
	srcDir := filepath.Join(absProjectPath, "vendor", "k8s.io", "kube-openapi")
	binDir := filepath.Join(absProjectPath, scaffold.BuildBinDir)

	if err := buildOpenAPIGenBinary(binDir, srcDir); err != nil {
		return err
	}

	gvMap, err := genutil.ParseGroupVersions()
	if err != nil {
		return fmt.Errorf("failed to parse group versions: (%v)", err)
	}
	gvb := &strings.Builder{}
	for g, vs := range gvMap {
		gvb.WriteString(fmt.Sprintf("%s:%v, ", g, vs))
	}

	log.Infof("Running OpenAPI code-generation for Custom Resource group versions: [%v]\n", gvb.String())

	apisPkg := filepath.Join(repoPkg, scaffold.ApisDir)
	fqApiStr := genutil.CreateFQApis(apisPkg, gvMap)
	fqApis := strings.Split(fqApiStr, ",")
	if err := openAPIGen(binDir, fqApis); err != nil {
		return err
	}

	s := &scaffold.Scaffold{}
	cfg := &input.Config{
		Repo:           repoPkg,
		AbsProjectPath: absProjectPath,
		ProjectName:    filepath.Base(absProjectPath),
	}
	crdMap, err := getCRDGVKMap()
	if err != nil {
		return err
	}
	for g, vs := range gvMap {
		for _, v := range vs {
			gvks := crdMap[filepath.Join(g, v)]
			for _, gvk := range gvks {
				r, err := scaffold.NewResource(filepath.Join(gvk.Group, gvk.Version), gvk.Kind)
				if err != nil {
					return err
				}
				err = s.Execute(cfg,
					&scaffold.CRD{Resource: r, IsOperatorGo: projutil.IsOperatorGo()},
				)
				if err != nil {
					return err
				}
			}
		}
	}

	log.Info("Code-generation complete.")
	return nil
}

func buildOpenAPIGenBinary(binDir, codegenSrcDir string) error {
	genDirs := []string{"./cmd/openapi-gen"}
	return genutil.BuildCodegenBinaries(genDirs, binDir, codegenSrcDir)
}

func openAPIGen(binDir string, fqApis []string) (err error) {
	if headerFile == "" {
		f, err := ioutil.TempFile(scaffold.BuildBinDir, "")
		if err != nil {
			return err
		}
		headerFile = f.Name()
		defer func() {
			if err = os.RemoveAll(headerFile); err != nil {
				log.Error(err)
			}
		}()
	}
	cgPath := filepath.Join(binDir, "openapi-gen")
	for _, fqApi := range fqApis {
		args := []string{
			"--input-dirs", fqApi,
			"--output-package", fqApi,
			"--output-file-base", "zz_generated.openapi",
			// openapi-gen requires a boilerplate file. Either use header or an
			// empty file if header is empty.
			"--go-header-file", headerFile,
		}
		cmd := exec.Command(cgPath, args...)
		if projutil.IsGoVerbose() {
			err = projutil.ExecCmd(cmd)
		} else {
			cmd.Stdout = ioutil.Discard
			cmd.Stderr = ioutil.Discard
			err = cmd.Run()
		}
		if err != nil {
			return fmt.Errorf("failed to perform openapi code-generation: %v", err)
		}
	}
	return nil
}

func getCRDGVKMap() (map[string][]metav1.GroupVersionKind, error) {
	crdInfos, err := ioutil.ReadDir(scaffold.CRDsDir)
	if err != nil {
		return nil, err
	}
	crdMap := make(map[string][]metav1.GroupVersionKind)
	for _, info := range crdInfos {
		if filepath.Ext(info.Name()) == ".yaml" {
			path := filepath.Join(scaffold.CRDsDir, info.Name())
			b, err := ioutil.ReadFile(path)
			if err != nil {
				return nil, err
			}
			crd := &apiextv1beta1.CustomResourceDefinition{}
			if err := yaml.Unmarshal(b, crd); err != nil {
				return nil, err
			}
			if crd.Kind != "CustomResourceDefinition" {
				continue
			}
			gv := filepath.Join(strings.Split(info.Name(), "_")[:2]...)
			crdMap[gv] = append(crdMap[gv], metav1.GroupVersionKind{
				Group:   crd.Spec.Group,
				Version: crd.Spec.Version,
				Kind:    crd.Spec.Names.Kind,
			})
		}
	}
	return crdMap, nil
}
