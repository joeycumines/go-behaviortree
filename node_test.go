/*
   Copyright 2019 Joseph Cumines

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package behaviortree

import (
	"github.com/go-test/deep"
	"strings"
	"testing"
)

func TestNode_Tick_nil(t *testing.T) {
	var (
		status Status
		err    error
	)

	var tree Node = nil

	status, err = tree.Tick()

	if status != Failure {
		t.Error("expected status to be failure but it was", status)
	}

	if err == nil {
		t.Error("expected non-nil error but it was", err)
	}
}

func TestNode_Tick_nilTick(t *testing.T) {
	var (
		status Status
		err    error
	)

	var tree Node = func() (Tick, []Node) {
		return nil, nil
	}

	status, err = tree.Tick()

	if status != Failure {
		t.Error("expected status to be failure but it was", status)
	}

	if err == nil {
		t.Error("expected non-nil error but it was", err)
	}
}

func TestNewNode(t *testing.T) {
	out := make(chan int)
	var (
		status Status
		err    error
	)

	tree := NewNode(
		Sequence,
		[]Node{
			NewNode(
				func(children []Node) (Status, error) {
					out <- 1
					return Success, nil
				},
				nil,
			),
			NewNode(
				func(children []Node) (Status, error) {
					out <- 2
					return Success, nil
				},
				nil,
			),
			NewNode(
				func(children []Node) (Status, error) {
					out <- 3
					return Success, nil
				},
				nil,
			),
		},
	)

	go func() {
		status, err = tree.Tick()
		out <- 4
	}()

	expected := []int{1, 2, 3, 4}
	actual := []int{
		<-out,
		<-out,
		<-out,
		<-out,
	}

	if diff := deep.Equal(expected, actual); diff != nil {
		t.Fatalf("expected tick order != actual: %s", strings.Join(diff, "\n  >"))
	}

	if status != Success {
		t.Error("expected status to be success but it was", status)
	}

	if err != nil {
		t.Error("expected nil error but it was", err)
	}
}
