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
	"testing"

	"github.com/coreos/operator-sdk/pkg/sdk/types"
	v1alpha1 "github.com/jmckind/rethinkdb-operator/pkg/apis/operator/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HandlerTestSuite struct {
	suite.Suite
}

func (suite *HandlerTestSuite) SetupTest() {
	// Run before each test...
}

func (suite *HandlerTestSuite) TestHandleWithNilObject() {
	context := types.Context{}
	event := types.Event{}
	assert.Nil(suite.T(), event.Object)

	handler := NewRethinkDBHandler()
	err := handler.Handle(context, event)
	assert.Nil(suite.T(), err)
}

func (suite *HandlerTestSuite) TestHandleWithDefaultRethinkDB() {
	context := types.Context{}
	event := types.Event{Object: &v1alpha1.RethinkDB{}}

	handler := NewRethinkDBHandler()
	err := handler.Handle(context, event)

	assert.Error(suite.T(), err)
}

// Run test suite...
func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
