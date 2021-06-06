/*
   Copyright 2021 Joseph Cumines

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
	"runtime"
	"testing"
	"time"
)

func TestNode_Value_race(t *testing.T) {
	defer checkNumGoroutines(t)(false, waitNumGoroutinesDefault*3)
	done := make(chan struct{})
	defer close(done)
	type k1 struct{}
	type k2 struct{}
	nodeOther := nn(Sequence).WithValue(k1{}, 5)
	for i := 0; i < 3000; i++ {
		node := nodeOther
		nodeOther = func() (Tick, []Node) { return node() }
		go func() {
			ticker := time.NewTicker(time.Millisecond * 10)
			defer ticker.Stop()
			for {
				node()
				select {
				case <-done:
					return
				case <-ticker.C:
				}
			}
		}()
	}
	go func() {
		ticker := time.NewTicker(time.Millisecond * 10)
		defer ticker.Stop()
		node := nn(Sequence).WithValue(k2{}, 3)
		for {
			node()
			select {
			case <-done:
				return
			case <-ticker.C:
			}
		}
	}()
	node := func() Node {
		node := nn(Sequence).WithValue(k1{}, 6)
		return func() (Tick, []Node) {
			time.Sleep(time.Millisecond * 100)
			return node()
		}
	}()
	n := time.Now()
	if v := node.Value(k1{}); v != 6 {
		t.Error(v)
	}
	if v := node.Value(k2{}); v != nil {
		t.Error(v)
	}
	t.Log(time.Since(n))
}

func nn(tick Tick, children ...Node) Node { return func() (Tick, []Node) { return tick, children } }

//noinspection GoNilness
func TestNode_Value_simple(t *testing.T) {
	type k1 struct{}
	type k2 struct{}
	if v := (Node)(nil).Value(k1{}); v != nil {
		t.Error(v)
	}
	if v := (Node)(nil).Value(k2{}); v != nil {
		t.Error(v)
	}
	n1 := nn(Sequence)
	if v := n1.Value(k1{}); v != nil {
		t.Error(v)
	}
	if v := n1.Value(k2{}); v != nil {
		t.Error(v)
	}
	n2 := n1.WithValue(k1{}, `v1`)
	if v := n2.Value(k1{}); v != `v1` {
		t.Error(v)
	}
	if frame, _ := runtime.CallersFrames(valueDataCaller[:]).Next(); frame.Function != `github.com/joeycumines/go-behaviortree.Node.valueSync` {
		t.Error(frame)
	}
	if v := n2.Value(k2{}); v != nil {
		t.Error(v)
	}
	n3 := n2.WithValue(k2{}, `v2`)
	if v := n3.Value(k1{}); v != `v1` {
		t.Error(v)
	}
	if v := n3.Value(k2{}); v != `v2` {
		t.Error(v)
	}
	n4 := n3.WithValue(k1{}, `v3`)
	if v := n4.Value(k1{}); v != `v3` {
		t.Error(v)
	}
	if v := n4.Value(k2{}); v != `v2` {
		t.Error(v)
	}
	if v := n3.Value(k1{}); v != `v1` {
		t.Error(v)
	}
	if v := n3.Value(k2{}); v != `v2` {
		t.Error(v)
	}
}

func TestNode_Value_noCaller(t *testing.T) {
	done := make(chan struct{})
	defer func() func() {
		old := runtimeCallers
		runtimeCallers = func(skip int, pc []uintptr) int {
			close(done)
			return 0
		}
		return func() {
			runtimeCallers = old
		}
	}()()
	if v := nn(Sequence).Value(nil); v != nil {
		t.Error(v)
	}
	select {
	case <-done:
	default:
		t.Error(`expected done`)
	}
	valueDataMutex.Lock()
	//lint:ignore SA2001 ensuring the lock can be acquired
	valueDataMutex.Unlock()
}

func TestNode_Value_nested(t *testing.T) {
	type k1 struct{}
	node := nn(Sequence).WithValue(k1{}, 5)
	if v := node.Value(k1{}); v != 5 {
		t.Fatal(v)
	}
	for i := 0; i < 3000; i++ {
		old := node
		node = func() (Tick, []Node) { return old() }
		if v := node.Value(k1{}); v != 5 {
			t.Fatal(i, v)
		}
	}
}

func TestNode_WithValue_panicNilReceiver(t *testing.T) {
	defer func() {
		if r := fmt.Sprint(recover()); r != `behaviortree.Node.WithValue nil receiver` {
			t.Error(r)
		}
	}()
	Node(nil).WithValue(1, 2)
	t.Error(`expected panic`)
}

func TestNode_WithValue_panicNilKey(t *testing.T) {
	defer func() {
		if r := fmt.Sprint(recover()); r != `behaviortree.Node.WithValue nil key` {
			t.Error(r)
		}
	}()
	Node(func() (Tick, []Node) { return nil, nil }).WithValue(nil, 2)
	t.Error(`expected panic`)
}

func TestNode_WithValue_panicNotComparible(t *testing.T) {
	defer func() {
		if r := fmt.Sprint(recover()); r != `behaviortree.Node.WithValue key is not comparable` {
			t.Error(r)
		}
	}()
	Node(func() (Tick, []Node) { return nil, nil }).WithValue([]int(nil), 2)
	t.Error(`expected panic`)
}

var Result interface{}

func benchTick(node Node) (status Status, err error) {
	for {
		status, err = node.Tick()
		if err != nil {
			return
		}
		if status == Failure {
			return
		}
	}
}

func Benchmark_newExampleCounter_withValue(b *testing.B) {
	var (
		status Status
		err    error
	)
	for i := 0; i < b.N; i++ {
		status, err = benchTick(newExampleCounter())
		if err != nil {
			b.Fatal(err)
		}
	}
	Result = status
}

func Benchmark_newExampleCounter_sansValue(b *testing.B) {
	b.StopTimer()
	defer func() func() {
		old := factory
		factory = func(tick Tick, children []Node) (node Node) {
			return func() (Tick, []Node) {
				return tick, children
			}
		}
		return func() {
			factory = old
		}
	}()()
	b.StartTimer()
	var (
		status Status
		err    error
	)
	for i := 0; i < b.N; i++ {
		status, err = benchTick(newExampleCounter())
		if err != nil {
			b.Fatal(err)
		}
	}
	Result = status
}

func Benchmark_newExampleCounter_withValueBackgroundStringer(b *testing.B) {
	{
		b.StopTimer()
		node := newExampleCounter()
		done := make(chan struct{})
		defer close(done)
		go func() {
			ticker := time.NewTicker(time.Millisecond)
			defer ticker.Stop()
			for {
				_ = node.String()
				select {
				case <-done:
					return
				case <-ticker.C:
				}
			}
		}()
		time.Sleep(time.Millisecond * 50)
		b.StartTimer()
	}
	var (
		status Status
		err    error
	)
	for i := 0; i < b.N; i++ {
		status, err = benchTick(newExampleCounter())
		if err != nil {
			b.Fatal(err)
		}
	}
	Result = status
}
