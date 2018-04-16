/*
   Copyright 2018 Joseph Cumines

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
	"testing"
	"reflect"
	"errors"
)

func TestSync_empty(t *testing.T) {
	nodes := make([]Node, 0)
	newNodes := Sync(nodes)
	if len(newNodes) != 0 {
		t.Error("bad nodes")
	}
}

func TestSync_set(t *testing.T) {
	nodes := make([]Node, 2)
	child := NewNode(Sequence, nil)
	called := false
	node := NewNode(func(children []Node) (Status, error) {
		if len(children) != 1 || reflect.ValueOf(child).Pointer() != reflect.ValueOf(children[0]).Pointer() {
			t.Fatal("bad child in node")
		}
		called = true
		return Success, errors.New("suk")
	}, []Node{child})
	nodes[1] = node
	newNodes := Sync(nodes)
	if len(newNodes) != 2 {
		t.Error("bad nodes")
	}
	if newNodes[0] != nil {
		t.Error("bad value")
	}
	nodes[0] = NewNode(Sequence, nil)
	if newNodes[0] != nil {
		t.Error("bad value")
	}
	if newNodes[1] == nil {
		t.Fatal("bad value")
	}
	if reflect.ValueOf(node).Pointer() == reflect.ValueOf(newNodes[1]).Pointer() {
		t.Fatal("bad value")
	}
	tick, children := newNodes[1]()
	if len(children) != 1 {
		t.Fatal("bad value")
	}
	status, err := tick(children)
	if !called {
		t.Fatal("bad call")
	}
	if status != Success || err == nil || err.Error() != "suk" {
		t.Fatal("bad tick value")
	}
}

func TestSync_nil(t *testing.T) {
	if nodes := Sync(nil); nodes != nil {
		t.Fatal("expected nil nodes")
	}
}

func TestSync_nilTick(t *testing.T) {
	var node Node = func() (Tick, []Node) {
		return nil, nil
	}
	nodes := Sync([]Node{
		func() (Tick, []Node) {
			return nil, []Node{node}
		},
	})
	if len(nodes) != 1 {
		t.Fatal("unexpected nodes", nodes)
	}
	tick, children := nodes[0]()
	if tick != nil {
		t.Error("expected a nil tick")
	}
	if len(children) != 1 || reflect.ValueOf(children[0]).Pointer() != reflect.ValueOf(node).Pointer() {
		t.Error("expected children to be returned", children)
	}
}

// TODO: finish testing sync
