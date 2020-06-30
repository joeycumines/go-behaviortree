/*
   Copyright 2020 Joseph Cumines

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
	"fmt"
	"github.com/go-test/deep"
	"strings"
	"testing"
)

func TestNode_Tick_nil(t *testing.T) {
	var (
		status Status
		err    error
	)

	var tree Node

	//noinspection GoNilness
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

func TestStatus_String(t *testing.T) {
	testCases := []struct {
		Status Status
		String string
	}{
		{
			Status: Success,
			String: `success`,
		},
		{
			Status: Failure,
			String: `failure`,
		},
		{
			Status: Running,
			String: `running`,
		},
		{
			Status: 0,
			String: `unknown status (0)`,
		},
		{
			Status: 234,
			String: `unknown status (234)`,
		},
	}

	for i, testCase := range testCases {
		name := fmt.Sprintf("TestStatus_String_#%d", i)

		if actual := testCase.Status.String(); actual != testCase.String {
			t.Errorf("%s failed: expected stringer '%s' != actual '%s'", name, testCase.String, actual)
		}
	}
}

func TestStatus_Status(t *testing.T) {
	testCases := []struct {
		Status   Status
		Expected Status
	}{
		{
			Status:   Success,
			Expected: Success,
		},
		{
			Status:   Failure,
			Expected: Failure,
		},
		{
			Status:   Running,
			Expected: Running,
		},
		{
			Status:   0,
			Expected: Failure,
		},
		{
			Status:   234,
			Expected: Failure,
		},
	}

	for i, testCase := range testCases {
		name := fmt.Sprintf("TestStatus_Status_#%d", i)

		if actual := testCase.Status.Status(); actual != testCase.Expected {
			t.Errorf("%s failed: expected behaviortree.Status.Status '%s' != actual '%s'", name, testCase.Status, actual)
		}
	}
}

func genTestNewFrame() *Frame { return New(Sequence).Frame() }

func TestNew_frame(t *testing.T) {
	var (
		expected = newFrame(genTestNewFrame)
		actual   = genTestNewFrame()
	)
	if expected == nil || actual == nil || expected.PC == 0 || actual.PC == 0 || expected.PC != expected.Entry || actual.PC == actual.Entry {
		t.Fatal(expected, actual)
	}
	actual.PC = actual.Entry
	if *expected != *actual {
		t.Errorf("expected != actual\nEXPECTED: %#v\nACTUAL: %#v", expected, actual)
	}
}

func genTestNewNodeFrame() *Frame { return NewNode(Sequence, nil).Frame() }

func TestNewNode_frame(t *testing.T) {
	var (
		expected = newFrame(genTestNewNodeFrame)
		actual   = genTestNewNodeFrame()
	)
	if expected == nil || actual == nil || expected.PC == 0 || actual.PC == 0 || expected.PC != expected.Entry || actual.PC == actual.Entry {
		t.Fatal(expected, actual)
	}
	actual.PC = actual.Entry
	if *expected != *actual {
		t.Errorf("expected != actual\nEXPECTED: %#v\nACTUAL: %#v", expected, actual)
	}
}

func Test_factory_nilFrame(t *testing.T) {
	if v := New(nil).Value(vkFrame{}); v == nil {
		t.Error(v)
	}
	if v := New(nil).Value(1); v != nil {
		t.Error(v)
	}
	if v := NewNode(nil, nil).Value(vkFrame{}); v == nil {
		t.Error(v)
	}
	if v := NewNode(nil, nil).Value(1); v != nil {
		t.Error(v)
	}
	defer func() func() {
		old := runtimeCallers
		runtimeCallers = func(skip int, pc []uintptr) int { return 0 }
		return func() {
			runtimeCallers = old
		}
	}()()
	if v := New(nil).Value(vkFrame{}); v != nil {
		t.Error(v)
	}
	if v := New(nil).Value(1); v != nil {
		t.Error(v)
	}
	if v := NewNode(nil, nil).Value(vkFrame{}); v != nil {
		t.Error(v)
	}
	if v := NewNode(nil, nil).Value(1); v != nil {
		t.Error(v)
	}
}
