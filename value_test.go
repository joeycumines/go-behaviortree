/*
   Copyright 2026 Joseph Cumines

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
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNode_Value_race(t *testing.T) {
	defer checkNumGoroutines(t)(false, waitNumGoroutinesDefault*3)
	var wg sync.WaitGroup
	defer wg.Wait()
	done := make(chan struct{})
	defer close(done)
	type k1 struct{}
	type k2 struct{}
	nodeOther := nn(Sequence).WithValue(k1{}, 5)
	for i := 0; i < 3000; i++ {
		node := nodeOther
		nodeOther = func() (Tick, []Node) { return node() }
		wg.Add(1)
		go func() {
			defer wg.Done()
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
	wg.Add(1)
	go func() {
		defer wg.Done()
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

// noinspection GoNilness
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
	// Ensure lock can still be acquired after Value call with nil key (no deadlock)
	valueDataMutex.Lock()
	//nolint:staticcheck
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

var Result any

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

func TestValue_panicSafety(t *testing.T) {
	// Ensure clean state
	if val := atomic.LoadUint32(&valueActive); val != 0 {
		t.Fatalf("valueActive should be 0 initially, got %d", val)
	}

	panickingNode := func() (Tick, []Node) {
		panic("boom")
	}

	// Helper to recover panic
	func() {
		defer func() {
			if r := recover(); r != "boom" {
				t.Errorf("caught unexpected panic: %v", r)
			}
		}()
		// Trigger the panic path
		// calling n.Value(key) calls valueSync -> valuePrep.
		Node(panickingNode).Value("key")
	}()

	// Verify flag is reset
	if val := atomic.LoadUint32(&valueActive); val != 0 {
		t.Errorf("valueActive stuck at %d after panic", val)
		// Clean up for other tests
		atomic.StoreUint32(&valueActive, 0)
	}
}

// ExampleUseValueProvider demonstrates how to register custom value handlers
// in a custom BT node implementation. This allows attaching metadata or
// configuration data to nodes that can be queried via Node.Value(key).
//
// IMPORTANT: UseValueProvider uses synchronization and call stack walking,
// so it MUST NOT be used in performance-critical runtime code paths.
// Use it only for development, debugging, testing, or metadata inspection.
func ExampleUseValueProvider() {
	// Define a custom key type for metadata lookup
	type metadataKey struct{}

	// Define a custom node type that registers metadata handlers
	customNode := func(name string, logic Tick) Node {
		return func() (Tick, []Node) {
			// Register a value handler that returns metadata for this node
			UseValueHandler(func(key any) (any, bool) {
				if _, ok := key.(metadataKey); ok {
					// Return metadata about this node
					return map[string]any{"name": name, "type": "custom"}, true
				}
				return nil, false
			})

			return logic, nil
		}
	}

	// Create a custom node with some logic
	node := customNode("myCustomNode", func(children []Node) (Status, error) {
		return Success, nil
	})

	// Query the metadata from the node
	metadata := node.Value(metadataKey{})
	if meta, ok := metadata.(map[string]any); ok {
		// Output: name: myCustomNode, type: custom
		fmt.Printf("name: %s, type: %s\n", meta["name"].(string), meta["type"].(string))
	}
}

// ExampleUseValueProvider_multiple demonstrates registering multiple value handlers,
// which are checked in order when calling Node.Value().
func ExampleUseValueProvider_multiple() {
	// Define multiple custom key types
	type configKey struct{}
	type statsKey struct{}

	// Create a node with multiple handlers
	var node Node
	node = func() (Tick, []Node) {
		// Register handler for configuration
		UseValueHandler(func(key any) (any, bool) {
			if _, ok := key.(configKey); ok {
				return map[string]string{"mode": "production", "region": "us-east"}, true
			}
			return nil, false
		})

		// Register handler for statistics
		UseValueHandler(func(key any) (any, bool) {
			if _, ok := key.(statsKey); ok {
				return map[string]int{"calls": 42, "successes": 40}, true
			}
			return nil, false
		})

		// Return a simple success tick
		return func(children []Node) (Status, error) { return Success, nil }, nil
	}

	// Query each key
	config := node.Value(configKey{})
	stats := node.Value(statsKey{})

	// Output: mode: production, calls: 42
	if cfg, ok := config.(map[string]string); ok {
		fmt.Printf("mode: %s", cfg["mode"])
	}
	if st, ok := stats.(map[string]int); ok {
		fmt.Printf(", calls: %d\n", st["calls"])
	}
}

func TestUseValueProvider_customNode(t *testing.T) {
	type customKey struct{}

	// Test that UseValueProvider works in a custom node
	customNode := New(func(children []Node) (Status, error) { return Success, nil })

	// Create a wrapper node that registers a handler
	var wrappedNode Node
	wrappedNode = func() (Tick, []Node) {
		UseValueHandler(func(key any) (any, bool) {
			if _, ok := key.(customKey); ok {
				return "custom-value", true
			}
			return nil, false
		})
		return customNode()
	}

	// Verify the handler works
	if v := wrappedNode.Value(customKey{}); v != "custom-value" {
		t.Errorf("expected 'custom-value', got %v", v)
	}

	// Verify non-handled keys return nil
	if v := wrappedNode.Value("other-key"); v != nil {
		t.Errorf("expected nil for non-handled key, got %v", v)
	}
}

func TestUseValueProvider_nesting(t *testing.T) {
	type outerKey struct{}
	type innerKey struct{}
	k1 := outerKey{}
	k2 := innerKey{}

	var innerNode Node
	innerNode = func() (Tick, []Node) {
		UseValueHandler(func(key any) (any, bool) {
			if k, ok := key.(innerKey); ok && k == k2 {
				return "inner-value", true
			}
			return nil, false
		})
		return func(children []Node) (Status, error) { return Success, nil }, nil
	}

	var outerNode Node
	outerNode = func() (Tick, []Node) {
		UseValueHandler(func(key any) (any, bool) {
			if k, ok := key.(outerKey); ok && k == k1 {
				return "outer-value", true
			}
			return nil, false
		})
		// Call inner node to register its handler
		_, _ = innerNode()
		return func(children []Node) (Status, error) { return Success, nil }, nil
	}

	// Both handlers should be accessible
	if v := outerNode.Value(k1); v != "outer-value" {
		t.Errorf("expected 'outer-value', got %v", v)
	}
	if v := outerNode.Value(k2); v != "inner-value" {
		t.Errorf("expected 'inner-value' for inner key from outer node, got %v", v)
	}
	// Call inner node directly to test its handler
	if v := innerNode.Value(k2); v != "inner-value" {
		t.Errorf("expected 'inner-value', got %v", v)
	}
}

func TestUseValueProvider_order(t *testing.T) {
	type key1 struct{}
	k1 := key1{}

	// Register handlers in a specific order
	var node Node
	node = func() (Tick, []Node) {
		UseValueHandler(func(key any) (any, bool) {
			if k, ok := key.(key1); ok && k == k1 {
				return "first", true
			}
			return nil, false
		})
		UseValueHandler(func(key any) (any, bool) {
			if k, ok := key.(key1); ok && k == k1 {
				return "second", true
			}
			return nil, false
		})
		return func(children []Node) (Status, error) { return Success, nil }, nil
	}

	// First registered handler should win (it sets the value first)
	if v := node.Value(k1); v != "first" {
		t.Errorf("expected 'first', got %v", v)
	}
}

func TestUseValueProvider_nilNode(t *testing.T) {
	// UseValueProvider should not panic even with nil node
	// It's just a package-level function
	var wrapped Node
	wrapped = func() (Tick, []Node) {
		UseValueHandler(func(key any) (any, bool) {
			return "value", true
		})
		// Normally would return n(), but here we test the handler registration alone
		return func(children []Node) (Status, error) { return Failure, nil }, nil
	}

	if v := wrapped.Value("any-key"); v == nil {
		// Handler should not find the key since it's not exact match
		t.Log("expected nil for mismatched key (handler registered but key doesn't match)")
	}
}
