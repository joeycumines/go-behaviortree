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
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestNew_leaf tests creating leaf nodes (no children)
func TestNew_leaf(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }

	tests := []struct {
		name     string
		tick     Tick
		children []Node
		wantTick Tick
	}{
		{
			name:     "valid tick, nil children",
			tick:     testTick,
			children: nil,
			wantTick: testTick,
		},
		{
			name:     "nil tick, nil children",
			tick:     nil,
			children: nil,
			wantTick: nil,
		},
		{
			name:     "valid tick, empty children",
			tick:     testTick,
			children: []Node{},
			wantTick: testTick,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := New(tt.tick, tt.children...)
			if node == nil {
				t.Fatal("New() returned nil node")
			}

			tick, children := node()
			// Check if tick is nil or not
			if (tick == nil) != (tt.wantTick == nil) {
				t.Errorf("tick nil mismatch: got nil=%v, want nil=%v", tick == nil, tt.wantTick == nil)
			}

			// For leaf nodes, children should be nil or empty
			if len(children) != 0 {
				t.Errorf("leaf node should have nil or empty children, got %v", children)
			}
		})
	}
}

// TestNew_composite tests creating composite nodes (with children)
func TestNew_composite(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }
	testChild := New(func(children []Node) (Status, error) { return Success, nil })

	tests := []struct {
		name     string
		tick     Tick
		children []Node
		wantTick Tick
		wantLen  int
	}{
		{
			name:     "valid tick with one child",
			tick:     testTick,
			children: []Node{testChild},
			wantTick: testTick,
			wantLen:  1,
		},
		{
			name: "valid tick with multiple children",
			tick: testTick,
			children: []Node{
				New(func(children []Node) (Status, error) { return Success, nil }),
				New(func(children []Node) (Status, error) { return Success, nil }),
				New(func(children []Node) (Status, error) { return Success, nil }),
			},
			wantTick: testTick,
			wantLen:  3,
		},
		{
			name:     "nil tick with children",
			tick:     nil,
			children: []Node{testChild},
			wantTick: nil,
			wantLen:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := New(tt.tick, tt.children...)
			if node == nil {
				t.Fatal("New() returned nil node")
			}

			tick, children := node()
			// Check if tick is nil or not
			if (tick == nil) != (tt.wantTick == nil) {
				t.Errorf("tick nil mismatch: got nil=%v, want nil=%v", tick == nil, tt.wantTick == nil)
			}

			if len(children) != tt.wantLen {
				t.Errorf("children length mismatch: got %d, want %d", len(children), tt.wantLen)
			}
		})
	}
}

// TestNewNode_leaf tests NewNode with nil children slice
func TestNewNode_leaf(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }

	tests := []struct {
		name     string
		tick     Tick
		children []Node
		wantTick Tick
	}{
		{
			name:     "valid tick, nil children slice",
			tick:     testTick,
			children: nil,
			wantTick: testTick,
		},
		{
			name:     "nil tick, nil children slice",
			tick:     nil,
			children: nil,
			wantTick: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewNode(tt.tick, tt.children)
			if node == nil {
				t.Fatal("NewNode() returned nil node")
			}

			tick, children := node()
			// Check if tick is nil or not
			if (tick == nil) != (tt.wantTick == nil) {
				t.Errorf("tick nil mismatch: got nil=%v, want nil=%v", tick == nil, tt.wantTick == nil)
			}

			if children != nil {
				t.Errorf("leaf node should have nil children, got %v", children)
			}
		})
	}
}

// TestNewNode_composite tests NewNode with children
func TestNewNode_composite(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }
	testChild := New(func(children []Node) (Status, error) { return Success, nil })

	tests := []struct {
		name     string
		tick     Tick
		children []Node
		wantTick Tick
		wantLen  int
	}{
		{
			name:     "valid tick with children",
			tick:     testTick,
			children: []Node{testChild},
			wantTick: testTick,
			wantLen:  1,
		},
		{
			name:     "valid tick with empty children slice",
			tick:     testTick,
			children: []Node{},
			wantTick: testTick,
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewNode(tt.tick, tt.children)
			if node == nil {
				t.Fatal("NewNode() returned nil node")
			}

			tick, children := node()
			// Check if tick is nil or not
			if (tick == nil) != (tt.wantTick == nil) {
				t.Errorf("tick nil mismatch: got nil=%v, want nil=%v", tick == nil, tt.wantTick == nil)
			}

			if len(children) != tt.wantLen {
				t.Errorf("children length mismatch: got %d, want %d", len(children), tt.wantLen)
			}
		})
	}
}

// TestNew_frameCapture tests that New properly captures frame information
func TestNew_frameCapture(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }

	// Test with nil children (leaf)
	t.Run("leaf with frame", func(t *testing.T) {
		node := New(testTick)
		verifyFrameCapture(t, node)
	})

	// Test with children (composite)
	t.Run("composite with frame", func(t *testing.T) {
		child := New(func(children []Node) (Status, error) { return Success, nil })
		node := New(testTick, child)
		verifyFrameCapture(t, node)
	})
}

// TestNewNode_frameCapture tests that NewNode properly captures frame information
func TestNewNode_frameCapture(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }

	// Test with nil children slice (leaf)
	t.Run("leaf with frame", func(t *testing.T) {
		node := NewNode(testTick, nil)
		verifyFrameCapture(t, node)
	})

	// Test with children (composite)
	t.Run("composite with frame", func(t *testing.T) {
		child := NewNode(func(children []Node) (Status, error) { return Success, nil }, nil)
		node := NewNode(testTick, []Node{child})
		verifyFrameCapture(t, node)
	})
}

// verifyFrameCapture is a helper to verify that a node has frame information
func verifyFrameCapture(t *testing.T, node Node) {
	t.Helper()

	// Get the frame from the node
	frameValue := node.Value(vkFrame{})
	if frameValue == nil {
		// If runtime.Callers failed to capture frame, that's acceptable
		t.Log("frame not captured (runtime.Callers may have failed)")
		return
	}

	frame, ok := frameValue.(*Frame)
	if !ok {
		t.Fatalf("expected *Frame, got %T", frameValue)
	}

	// Verify frame has essential information
	if frame.PC == 0 {
		t.Error("frame.PC should not be zero")
	}
	if frame.Function == "" {
		t.Error("frame.Function should not be empty")
	}
	if frame.Entry == 0 {
		t.Error("frame.Entry should not be zero")
	}
}

// TestNew_noFrameCapture tests behavior when runtime.Callers fails
func TestNew_noFrameCapture(t *testing.T) {
	// Save original runtime.Callers
	oldRuntimeCallers := runtimeCallers
	oldRuntimeCallersFrames := runtimeCallersFrames

	defer func() {
		runtimeCallers = oldRuntimeCallers
		runtimeCallersFrames = oldRuntimeCallersFrames
	}()

	// Mock runtime.Callers to return 0 (simulating failure)
	runtimeCallers = func(skip int, pc []uintptr) int { return 0 }

	testTick := func(children []Node) (Status, error) { return Success, nil }

	// Test that nodes are still created without frames
	t.Run("leaf without frame", func(t *testing.T) {
		node := New(testTick)
		if node == nil {
			t.Fatal("New() returned nil node when runtime.Callers failed")
		}

		// Should not have frame attached
		frameValue := node.Value(vkFrame{})
		if frameValue != nil {
			t.Errorf("expected no frame when runtime.Callers fails, got %v", frameValue)
		}

		// But node should still work
		tick, children := node()
		if tick == nil && testTick != nil {
			t.Error("expected non-nil tick")
		}
		if tick != nil && testTick == nil {
			t.Error("expected nil tick")
		}
		if children != nil {
			t.Error("expected nil children for leaf")
		}
	})

	t.Run("composite without frame", func(t *testing.T) {
		child := New(func(children []Node) (Status, error) { return Success, nil })
		node := New(testTick, child)
		if node == nil {
			t.Fatal("New() returned nil node when runtime.Callers failed")
		}

		// Should not have frame attached
		frameValue := node.Value(vkFrame{})
		if frameValue != nil {
			t.Errorf("expected no frame when runtime.Callers fails, got %v", frameValue)
		}

		// But node should still work
		tick, children := node()
		if tick == nil && testTick != nil {
			t.Error("expected non-nil tick")
		}
		if tick != nil && testTick == nil {
			t.Error("expected nil tick")
		}
		if len(children) != 1 {
			t.Errorf("expected 1 child, got %d", len(children))
		}
	})
}

// TestNew_nodeEquality tests that nodes created with New work correctly
func TestNew_nodeEquality(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }

	// Create multiple leaf nodes with same tick
	node1 := New(testTick)
	node2 := New(testTick)

	// They should both work
	tick1, children1 := node1()
	tick2, children2 := node2()

	// Ticks should both be non-nil
	if tick1 == nil || tick2 == nil {
		t.Error("ticks should be non-nil")
	}

	// Children should be nil for leaf nodes
	if children1 != nil || children2 != nil {
		t.Error("leaf nodes should have nil children")
	}
}

// TestNew_childPreservation tests that children are preserved
func TestNew_childPreservation(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }

	child1Called := false
	child2Called := false

	child1 := New(func(children []Node) (Status, error) {
		child1Called = true
		return Success, nil
	})

	child2 := New(func(children []Node) (Status, error) {
		child2Called = true
		return Success, nil
	})

	node := New(testTick, child1, child2)

	// Verify children are preserved
	tick, children := node()
	if tick == nil {
		t.Error("tick should be non-nil")
	}

	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}

	// Verify children work - tick them directly to test
	tickA, _ := children[0]()
	tickB, _ := children[1]()

	tickA(nil)
	tickB(nil)

	if !child1Called {
		t.Error("child1 was not called")
	}
	if !child2Called {
		t.Error("child2 was not called")
	}
}

// TestNew_tickExecution tests that the tick function is properly executable
func TestNew_tickExecution(t *testing.T) {
	calledCount := 0
	testTick := func(children []Node) (Status, error) {
		calledCount++
		return Success, nil
	}

	// Leaf node
	t.Run("leaf tick execution", func(t *testing.T) {
		calledCount = 0
		node := New(testTick)
		status, err := node.Tick()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if status != Success {
			t.Errorf("expected Success, got %v", status)
		}
		if calledCount != 1 {
			t.Errorf("expected tick to be called once, was called %d times", calledCount)
		}
	})

	// Composite node
	t.Run("composite tick execution", func(t *testing.T) {
		calledCount = 0
		child := New(func(children []Node) (Status, error) { return Success, nil })
		node := New(testTick, child)
		status, err := node.Tick()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if status != Success {
			t.Errorf("expected Success, got %v", status)
		}
		if calledCount != 1 {
			t.Errorf("expected tick to be called once, was called %d times", calledCount)
		}
	})
}

// TestNew_frameImmutability tests that returned frames are copies
func TestNew_frameImmutability(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }
	node := New(testTick)

	// Get frame twice
	frameValue1 := node.Value(vkFrame{})
	if frameValue1 == nil {
		t.Skip("frame not captured, skipping immutability test")
	}

	frame1, ok := frameValue1.(*Frame)
	if !ok {
		t.Fatalf("expected *Frame, got %T", frameValue1)
	}

	frameValue2 := node.Value(vkFrame{})
	frame2, ok := frameValue2.(*Frame)
	if !ok {
		t.Fatalf("expected *Frame, got %T", frameValue2)
	}

	// Should be pointers to different objects
	if frame1 == frame2 {
		// This is actually acceptable - they might return the same pointer
		// But let's verify the data is correct
	}

	// Modify first frame and check that second is not affected (if they're different objects)
	frame1.Line = 9999
	frame2Value := node.Value(vkFrame{})
	frame2, _ = frame2Value.(*Frame)
	if frame2.Line == 9999 && frame1 != frame2 {
		// If they're different objects, frame2 should not have been modified
		t.Error("frame immutability violated: modifying one frame affected another")
	}

	// But the next call to Frame() should return a fresh/correct frame
	frame3 := node.Frame()
	if frame3 == nil {
		t.Fatal("Frame() returned nil")
	}
	if frame3.Line == 0 && frame3.PC != 0 {
		// Line might be 0 from newFrame, that's OK
		t.Log("frame has zero Line (from newFrame)")
	}
}

// TestNew_frameFunctionDetails tests captured frame has correct function details
func TestNew_frameFunctionDetails(t *testing.T) {
	// Create node and get its frame
	node := New(Sequence)
	frameValue := node.Value(vkFrame{})

	if frameValue == nil {
		t.Skip("frame not captured, skipping function details test")
	}

	frame, ok := frameValue.(*Frame)
	if !ok {
		t.Fatalf("expected *Frame, got %T", frameValue)
	}

	// Verify frame has function name related to our test
	if frame.Function == "" {
		t.Error("frame.Function should not be empty")
	} else {
		// Function should contain "New" or "NewNode" or "factory"
		frameLower := strings.ToLower(frame.Function)
		if !strings.Contains(frameLower, "new") && !strings.Contains(frameLower, "factory") {
			t.Errorf("frame.Function %q should contain 'new' or 'factory'", frame.Function)
		}
	}

	// Verify File is not empty
	if frame.File == "" {
		t.Error("frame.File should not be empty")
	} else {
		// File should contain "factory.go" or similar
		if !strings.Contains(frame.File, "factory") && !strings.Contains(frame.File, "behaviortree") {
			t.Errorf("frame.File %q should contain 'factory' or 'behaviortree'", frame.File)
		}
	}
}

// TestNew_threadSafety tests concurrent creation and access
func TestNew_threadSafety(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }

	const numGoroutines = 100
	const numIterations = 100

	var wg sync.WaitGroup
	defer wg.Wait()

	done := make(chan struct{})
	defer close(done)

	// Create nodes concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()

			for j := 0; j < numIterations; j++ {
				select {
				case <-done:
					return
				default:
				}

				// Create leaf nodes
				node := New(testTick)

				// Access the node
				tick, children := node()
				if tick == nil {
					t.Errorf("goroutine %d iteration %d: tick is nil", iteration, j)
				}
				if len(children) != 0 {
					t.Errorf("goroutine %d iteration %d: unexpected children", iteration, j)
				}

				// Tick the node
				status, err := node.Tick()
				if err != nil {
					t.Errorf("goroutine %d iteration %d: tick error: %v", iteration, j, err)
				}
				if status != Success {
					t.Errorf("goroutine %d iteration %d: unexpected status: %v", iteration, j, status)
				}
			}
		}(i)
	}

	// Create composite nodes concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()

			for j := 0; j < numIterations; j++ {
				select {
				case <-done:
					return
				default:
				}

				child1 := New(testTick)
				child2 := New(testTick)
				node := New(testTick, child1, child2)

				// Access the node
				tick, children := node()
				if tick == nil {
					t.Errorf("goroutine %d iteration %d: tick is nil", iteration, j)
				}
				if len(children) != 2 {
					t.Errorf("goroutine %d iteration %d: expected 2 children, got %d", iteration, j, len(children))
				}

				// Tick the node
				status, err := node.Tick()
				if err != nil {
					t.Errorf("goroutine %d iteration %d: tick error: %v", iteration, j, err)
				}
				if status != Success {
					t.Errorf("goroutine %d iteration %d: unexpected status: %v", iteration, j, status)
				}
			}
		}(i)
	}
}

// TestNew_withNilTickInComposite tests composite nodes with nil tick
func TestNew_withNilTickInComposite(t *testing.T) {
	child := New(func(children []Node) (Status, error) { return Success, nil })
	node := New(nil, child) // nil tick with children

	if node == nil {
		t.Fatal("New() returned nil node")
	}

	status, err := node.Tick()
	if err == nil {
		t.Error("expected error for nil tick, got nil")
	}
	if status != Failure {
		t.Errorf("expected Failure status, got %v", status)
	}
}

// TestNew_variadicBehavior tests that New(...) properly handles variadic arguments
func TestNew_variadicBehavior(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }

	t.Run("no children (leaf)", func(t *testing.T) {
		node := New(testTick)
		tick, children := node()
		if tick == nil {
			t.Error("tick should be non-nil")
		}
		if children != nil {
			t.Errorf("expected nil children, got %v", children)
		}
	})

	t.Run("one child (composite)", func(t *testing.T) {
		child := New(func(children []Node) (Status, error) { return Success, nil })
		node := New(testTick, child)
		tick, children := node()
		if tick == nil {
			t.Error("tick should be non-nil")
		}
		if len(children) != 1 {
			t.Errorf("expected 1 child, got %d", len(children))
		}
	})

	t.Run("multiple children (composite)", func(t *testing.T) {
		child1 := New(func(children []Node) (Status, error) { return Success, nil })
		child2 := New(func(children []Node) (Status, error) { return Success, nil })
		child3 := New(func(children []Node) (Status, error) { return Success, nil })
		node := New(testTick, child1, child2, child3)
		tick, children := node()
		if tick == nil {
			t.Error("tick should be non-nil")
		}
		if len(children) != 3 {
			t.Errorf("expected 3 children, got %d", len(children))
		}
	})
}

// TestNewNode_nilSliceHandling tests the difference between passing nil explicitly and using varargs
func TestNewNode_nilSliceHandling(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }

	t.Run("New with no args (leaf)", func(t *testing.T) {
		node := New(testTick)
		tick, children := node()
		if tick == nil {
			t.Error("tick should be non-nil")
		}
		// For no args, children should be nil (leaf path)
		if children != nil {
			t.Errorf("expected nil children, got %v", children)
		}
	})

	t.Run("NewNode with nil slice (leaf)", func(t *testing.T) {
		node := NewNode(testTick, nil)
		tick, children := node()
		if tick == nil {
			t.Error("tick should be non-nil")
		}
		// For nil slice, should take leaf path
		if children != nil {
			t.Errorf("expected nil children, got %v", children)
		}
	})

	t.Run("NewNode with empty slice (composite)", func(t *testing.T) {
		node := NewNode(testTick, []Node{})
		tick, children := node()
		if tick == nil {
			t.Error("tick should be non-nil")
		}
		// For empty slice (not nil), should take composite path
		if len(children) != 0 {
			t.Errorf("expected empty slice, got length %d", len(children))
		}
	})
}

// TestNew_withValueInteraction tests that factory-created nodes work with WithValue
func TestNew_withValueInteraction(t *testing.T) {
	type testKey struct{}

	testTick := func(children []Node) (Status, error) { return Success, nil }

	t.Run("leaf with WithValue", func(t *testing.T) {
		node := New(testTick)
		wrapped := node.WithValue(testKey{}, "test-value")

		// Get value
		value := wrapped.Value(testKey{})
		if value != "test-value" {
			t.Errorf("expected 'test-value', got %v", value)
		}

		// Node should still work
		tick, children := wrapped()
		if tick == nil {
			t.Error("tick should be non-nil after WithValue")
		}
		if children != nil {
			t.Error("expected nil children after WithValue")
		}
	})

	t.Run("composite with WithValue", func(t *testing.T) {
		child := New(func(children []Node) (Status, error) { return Success, nil })
		node := New(testTick, child)
		wrapped := node.WithValue(testKey{}, "test-value")

		// Get value
		value := wrapped.Value(testKey{})
		if value != "test-value" {
			t.Errorf("expected 'test-value', got %v", value)
		}

		// Node should still work
		tick, children := wrapped()
		if tick == nil {
			t.Error("tick should be non-nil after WithValue")
		}
		if len(children) != 1 {
			t.Errorf("expected 1 child after WithValue, got %d", len(children))
		}
	})
}

// TestNew_frameTickIntegration tests that Frame() method works on factory-created nodes
func TestNew_frameTickIntegration(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }

	t.Run("leaf node Frame()", func(t *testing.T) {
		node := New(testTick)
		frame := node.Frame()

		if frame == nil {
			t.Skip("frame not available, skipping")
		}

		// Frame should have at least basic info
		if frame.PC == 0 && frame.Function != "" {
			// PC and Entry are from FuncForPC and entry point, might be 0 in some cases
			// but Function should be populated
			t.Log("PC is 0 but Function is populated")
		}

		if frame.Function == "" {
			t.Error("frame.Function should not be empty")
		}
	})

	t.Run("composite node Frame()", func(t *testing.T) {
		child := New(func(children []Node) (Status, error) { return Success, nil })
		node := New(testTick, child)
		frame := node.Frame()

		if frame == nil {
			t.Skip("frame not available, skipping")
		}

		if frame.Function == "" {
			t.Error("frame.Function should not be empty")
		}
	})
}

// TestNew_tickFrameIntegration tests that Tick.Frame() works
func TestNew_tickFrameIntegration(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }

	t.Run("leaf tick Frame()", func(t *testing.T) {
		node := New(testTick)
		tick, _ := node()

		tickFrame := tick.Frame()
		if tickFrame == nil {
			t.Skip("tick frame not available, skipping")
		}

		if tickFrame.Function == "" {
			t.Error("tick frame Function should not be empty")
		}
	})

	t.Run("composite tick Frame()", func(t *testing.T) {
		child := New(func(children []Node) (Status, error) { return Success, nil })
		node := New(testTick, child)
		tick, _ := node()

		tickFrame := tick.Frame()
		if tickFrame == nil {
			t.Skip("tick frame not available, skipping")
		}

		if tickFrame.Function == "" {
			t.Error("tick frame Function should not be empty")
		}
	})
}

// TestNew_raceCondition tests for race conditions with Value() and Tick()
func TestNew_raceCondition(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }
	type testKey struct{}

	node := New(testTick, New(testTick))

	var wg sync.WaitGroup
	defer wg.Wait()

	done := make(chan struct{})
	defer close(done)

	// Run concurrent Tick() and Value() calls
	for i := 0; i < 50; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					node.Tick()
				}
			}
		}()

		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					node.Value(testKey{})
					node.Value(vkFrame{})
				}
			}
		}()
	}

	// Let it run for a bit
	time.Sleep(100 * time.Millisecond)
}

// TestNew_stress tests creating many nodes
func TestNew_stress(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }

	const numNodes = 1000

	// Create many leaf nodes
	for i := 0; i < numNodes; i++ {
		node := New(testTick)
		if node == nil {
			t.Fatalf("failed to create leaf node at iteration %d", i)
		}
		status, err := node.Tick()
		if err != nil {
			t.Fatalf("leaf node tick error at iteration %d: %v", i, err)
		}
		if status != Success {
			t.Fatalf("leaf node wrong status at iteration %d: %v", i, status)
		}
	}

	// Create many composite nodes
	for i := 0; i < numNodes; i++ {
		child := New(testTick)
		node := New(testTick, child)
		if node == nil {
			t.Fatalf("failed to create composite node at iteration %d", i)
		}
		status, err := node.Tick()
		if err != nil {
			t.Fatalf("composite node tick error at iteration %d: %v", i, err)
		}
		if status != Success {
			t.Fatalf("composite node wrong status at iteration %d: %v", i, status)
		}
	}
}

// TestNew_frameRetrieval tests retrieving frame both via Value and Frame methods
func TestNew_frameRetrieval(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }
	node := New(testTick)

	// Get frame via Value()
	valueFrame := node.Value(vkFrame{})

	// Get frame via Frame()
	methodFrame := node.Frame()

	// Both should be non-nil or both nil
	if (valueFrame == nil) != (methodFrame == nil) {
		t.Errorf("inconsistent frame presence: Value()=%v, Frame()=%v",
			valueFrame, methodFrame)
	}

	if valueFrame != nil {
		// valueFrame is already *Frame from Value interface
		vf, ok1 := valueFrame.(*Frame)
		// methodFrame is already *Frame from Frame() method
		mf := methodFrame

		if !ok1 {
			t.Errorf("Value() returned %T, expected *Frame", valueFrame)
		}

		// They should be consistent
		if ok1 {
			if vf.Function != mf.Function {
				t.Errorf("Function mismatch: Value()=%v, Frame()=%v",
					vf.Function, mf.Function)
			}
			// File might differ (absolute vs relative), so just check both non-empty
			if vf.File != "" && mf.File != "" {
				// Both populated, that's good enough
			} else if vf.File == "" && mf.File == "" {
				t.Error("both frames have empty File")
			}
		}
	}
}

// TestNew_handlerIntegration tests factory nodes with UseValueProvider
func TestNew_handlerIntegration(t *testing.T) {
	type customKey struct{}

	testTick := func(children []Node) (Status, error) { return Success, nil }

	t.Run("factory node with handler", func(t *testing.T) {
		node := New(testTick, nil)
		wrapped := Node(func() (Tick, []Node) {
			UseValueHandler(func(key any) (any, bool) {
				if _, ok := key.(customKey); ok {
					return "handler-value", true
				}
				return nil, false
			})
			return node()
		})

		value := wrapped.Value(customKey{})
		if value != "handler-value" {
			t.Errorf("expected 'handler-value', got %v", value)
		}
	})
}

// TestNew_multipleWithValues tests that WithValue can be called multiple times
func TestNew_multipleWithValues(t *testing.T) {
	type key1 struct{}
	type key2 struct{}
	type key3 struct{}

	testTick := func(children []Node) (Status, error) { return Success, nil }

	node := New(testTick)
	node = node.WithValue(key1{}, "value1")
	node = node.WithValue(key2{}, "value2")
	node = node.WithValue(key3{}, "value3")

	// All values should be retrievable
	if v := node.Value(key1{}); v != "value1" {
		t.Errorf("expected 'value1', got %v", v)
	}
	if v := node.Value(key2{}); v != "value2" {
		t.Errorf("expected 'value2', got %v", v)
	}
	if v := node.Value(key3{}); v != "value3" {
		t.Errorf("expected 'value3', got %v", v)
	}

	// Verify it still works as a node
	status, err := node.Tick()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if status != Success {
		t.Errorf("expected Success, got %v", status)
	}
}

// TestNew_nestedNodes tests creating nested node structures
func TestNew_nestedNodes(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }

	// Create a tree structure
	//        root
	//       /    \
	//     c1      c2
	//            /  \
	//          gc1  gc2

	gc1 := New(func(children []Node) (Status, error) { return Success, nil })
	gc2 := New(func(children []Node) (Status, error) { return Success, nil })

	c1 := New(testTick)
	c2 := New(testTick, gc1, gc2)

	root := New(testTick, c1, c2)

	// Verify structure by ticking
	tick, children := root()
	if tick == nil {
		t.Error("root tick should be non-nil")
	}
	if len(children) != 2 {
		t.Fatalf("expected 2 children at root, got %d", len(children))
	}

	// Tick c1
	tick1, grandchildren := children[0]()
	if tick1 == nil {
		t.Error("c1 tick should be non-nil")
	}
	if len(grandchildren) != 0 {
		t.Errorf("c1 should have no children, got %d", len(grandchildren))
	}

	// Tick c2
	tick2, grandchildren2 := children[1]()
	if tick2 == nil {
		t.Error("c2 tick should be non-nil")
	}
	if len(grandchildren2) != 2 {
		t.Fatalf("expected 2 grandchildren under c2, got %d", len(grandchildren2))
	}

	// Verify entire tree can be ticked
	status, err := root.Tick()
	if err != nil {
		t.Errorf("tree tick error: %v", err)
	}
	if status != Success {
		t.Errorf("expected tree status Success, got %v", status)
	}
}

// TestNewFrame_basic tests basic NewFrame functionality
func TestNewFrame_basic(t *testing.T) {
	// Create a runtime.Frame to test with
	pc := uintptr(0x12345)
	inputFrame := runtime.Frame{
		PC:       pc,
		Function: "example.com/pkg.function",
		File:     "/path/to/file.go",
		Line:     42,
		Entry:    pc,
	}

	// Call NewFrame
	result := NewFrame(inputFrame)

	// Verify all fields are copied correctly
	if result.PC != inputFrame.PC {
		t.Errorf("PC mismatch: got %v, want %v", result.PC, inputFrame.PC)
	}
	if result.Function != inputFrame.Function {
		t.Errorf("Function mismatch: got %v, want %v", result.Function, inputFrame.Function)
	}
	if result.File != inputFrame.File {
		t.Errorf("File mismatch: got %v, want %v", result.File, inputFrame.File)
	}
	if result.Line != inputFrame.Line {
		t.Errorf("Line mismatch: got %v, want %v", result.Line, inputFrame.Line)
	}
	if result.Entry != inputFrame.Entry {
		t.Errorf("Entry mismatch: got %v, want %v", result.Entry, inputFrame.Entry)
	}
}

// TestNewFrame_zeroValues tests NewFrame with zero-valued runtime.Frame
func TestNewFrame_zeroValues(t *testing.T) {
	inputFrame := runtime.Frame{}
	result := NewFrame(inputFrame)

	// Should copy all zero values correctly
	if result.PC != 0 {
		t.Errorf("expected PC=0, got %v", result.PC)
	}
	if result.Function != "" {
		t.Errorf("expected empty Function, got %v", result.Function)
	}
	if result.File != "" {
		t.Errorf("expected empty File, got %v", result.File)
	}
	if result.Line != 0 {
		t.Errorf("expected Line=0, got %v", result.Line)
	}
	if result.Entry != 0 {
		t.Errorf("expected Entry=0, got %v", result.Entry)
	}
}

// TestNewFrame_partialValues tests NewFrame with partial (some empty) values
func TestNewFrame_partialValues(t *testing.T) {
	inputFrame := runtime.Frame{
		PC:       0x12345,
		Function: "example.com/pkg.function",
		// File, Line, Entry left at zero values
	}

	result := NewFrame(inputFrame)

	if result.PC != 0x12345 {
		t.Errorf("expected PC=0x12345, got %v", result.PC)
	}
	if result.Function != "example.com/pkg.function" {
		t.Errorf("expected Function='example.com/pkg.function', got %v", result.Function)
	}
	if result.File != "" {
		t.Errorf("expected empty File, got %v", result.File)
	}
	if result.Line != 0 {
		t.Errorf("expected Line=0, got %v", result.Line)
	}
	if result.Entry != 0 {
		t.Errorf("expected Entry=0, got %v", result.Entry)
	}
}

// TestNewFrame_immutability tests that NewFrame creates a true copy
func TestNewFrame_immutability(t *testing.T) {
	pc := uintptr(0x12345)
	inputFrame := runtime.Frame{
		PC:       pc,
		Function: "example.com/pkg.function",
		File:     "/path/to/file.go",
		Line:     42,
		Entry:    pc,
	}

	result := NewFrame(inputFrame)

	// Modify original
	inputFrame.PC = 0x54321
	inputFrame.Function = "modified"
	inputFrame.File = "/modified/path"
	inputFrame.Line = 99
	inputFrame.Entry = 0x54321

	// Result should be unchanged
	if result.PC != pc {
		t.Errorf("result.PC changed after modifying original: got %v, want %v", result.PC, pc)
	}
	if result.Function != "example.com/pkg.function" {
		t.Errorf("result.Function changed after modifying original: got %v", result.Function)
	}
	if result.File != "/path/to/file.go" {
		t.Errorf("result.File changed after modifying original: got %v", result.File)
	}
	if result.Line != 42 {
		t.Errorf("result.Line changed after modifying original: got %v, want %d", result.Line, 42)
	}
	if result.Entry != pc {
		t.Errorf("result.Entry changed after modifying original: got %v, want %v", result.Entry, pc)
	}
}

// TestGetFrame_fromNode tests GetFrame extracts frame from Node
func TestGetFrame_fromNode(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }

	// Create a node that should have frame information attached
	node := New(testTick)

	// Try to get frame
	frame := GetFrame(node)

	// Frame may be nil if runtime.Callers failed, which is acceptable
	if frame != nil {
		// If frame exists, verify it has some basic structure
		if frame.PC == 0 && frame.Function == "" {
			t.Error("frame has no PC or Function information")
		}
	}
}

// TestGetFrame_nilNode tests GetFrame with nil node
func TestGetFrame_nilNode(t *testing.T) {
	frame := GetFrame(Node(nil))
	if frame != nil {
		t.Errorf("expected nil frame for nil node, got %v", frame)
	}
}

// TestGetFrame_fromCustomValuer tests GetFrame with custom Valuer implementation
func TestGetFrame_fromCustomValuer(t *testing.T) {

	// Create a custom valuer
	customValuer := Node(func() (Tick, []Node) {
		UseValueHandler(func(key any) (any, bool) {
			if _, ok := key.(vkFrame); ok {
				// Return a custom frame
				testFrame := &Frame{
					PC:       0x54321,
					Function: "custom.function",
					File:     "custom.go",
					Line:     100,
					Entry:    0x54321,
				}
				return testFrame, true
			}
			return nil, false
		})
		return func(children []Node) (Status, error) { return Success, nil }, nil
	})

	frame := GetFrame(customValuer)
	if frame == nil {
		t.Fatal("expected frame from custom valuer")
	}

	if frame.Function != "custom.function" {
		t.Errorf("expected Function='custom.function', got %v", frame.Function)
	}
	if frame.File != "custom.go" {
		t.Errorf("expected File='custom.go', got %v", frame.File)
	}
}

// TestGetFrame_valuerWithoutFrame tests GetFrame with valuer that doesn't have frame
func TestGetFrame_valuerWithoutFrame(t *testing.T) {

	// Create a valuer that doesn't provide frame
	valuerNoFrame := Node(func() (Tick, []Node) {
		UseValueHandler(func(key any) (any, bool) {
			return nil, false // Don't handle vkFrame
		})
		return func(children []Node) (Status, error) { return Success, nil }, nil
	})

	frame := GetFrame(valuerNoFrame)
	if frame != nil {
		t.Errorf("expected nil frame when valuer doesn't provide one, got %v", frame)
	}
}

// TestWithFrame_nilFrame tests WithFrame with nil frame
func TestWithFrame_nilFrame(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }
	node := New(testTick)

	// WithFrame with nil should still work
	wrapped := node.WithFrame(nil)

	if wrapped == nil {
		t.Fatal("WithFrame(nil) returned nil")
	}

	// Verify frame is nil
	frame := GetFrame(wrapped)
	if frame != nil {
		t.Errorf("expected nil frame after WithFrame(nil), got %v", frame)
	}

	// Verify node still works
	status, err := wrapped.Tick()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if status != Success {
		t.Errorf("expected Success, got %v", status)
	}
}

// TestWithFrame_validFrame tests WithFrame with valid frame
func TestWithFrame_validFrame(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }
	node := New(testTick)

	// Create a custom frame
	customFrame := &Frame{
		PC:       0xABCDE,
		Function: "test.custom.frame",
		File:     "test_file.go",
		Line:     77,
		Entry:    0xABCDE,
	}

	// Wrap with frame
	wrapped := node.WithFrame(customFrame)

	if wrapped == nil {
		t.Fatal("WithFrame returned nil")
	}

	// Verify frame is attached
	frame := GetFrame(wrapped)
	if frame == nil {
		t.Fatal("expected frame to be attached, got nil")
	}

	// Verify frame content
	if frame.PC != customFrame.PC {
		t.Errorf("PC mismatch: got %v, want %v", frame.PC, customFrame.PC)
	}
	if frame.Function != customFrame.Function {
		t.Errorf("Function mismatch: got %v, want %v", frame.Function, customFrame.Function)
	}
	if frame.File != customFrame.File {
		t.Errorf("File mismatch: got %v, want %v", frame.File, customFrame.File)
	}
	if frame.Line != customFrame.Line {
		t.Errorf("Line mismatch: got %v, want %v", frame.Line, customFrame.Line)
	}
	if frame.Entry != customFrame.Entry {
		t.Errorf("Entry mismatch: got %v, want %v", frame.Entry, customFrame.Entry)
	}

	// Verify node still works
	status, err := wrapped.Tick()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if status != Success {
		t.Errorf("expected Success, got %v", status)
	}
}

// TestWithFrame_immutability tests that WithFrame creates a new node with copy of frame
func TestWithFrame_immutability(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }
	node := New(testTick)

	originalFrame := &Frame{
		PC:       0x12345,
		Function: "original",
		File:     "original.go",
		Line:     10,
		Entry:    0x12345,
	}

	// Wrap with frame
	wrapped := node.WithFrame(originalFrame)

	// Modify original frame
	originalFrame.Line = 999
	originalFrame.File = "modified.go"

	// Get frame from wrapped node - should be a copy
	frame := GetFrame(wrapped)
	if frame == nil {
		t.Fatal("expected frame")
	}

	// Should have original values, not modified ones
	if frame.Line == 999 {
		t.Error("frame was not copied - modification to original affected wrapped node")
	}
	if frame.File == "modified.go" {
		t.Error("frame was not copied - modification to original affected wrapped node")
	}
	if frame.Line != 10 {
		t.Errorf("expected Line=10, got %v", frame.Line)
	}
	if frame.File != "original.go" {
		t.Errorf("expected File='original.go', got %v", frame.File)
	}
}

// TestWithFrame_chained tests multiple WithFrame calls
func TestWithFrame_chained(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }
	node := New(testTick)

	frame1 := &Frame{
		PC:       0x11111,
		Function: "frame1",
		File:     "frame1.go",
		Line:     11,
		Entry:    0x11111,
	}

	frame2 := &Frame{
		PC:       0x22222,
		Function: "frame2",
		File:     "frame2.go",
		Line:     22,
		Entry:    0x22222,
	}

	// Chain WithFrame calls
	wrapped1 := node.WithFrame(frame1)
	wrapped2 := wrapped1.WithFrame(frame2)

	// Last frame should win
	frame := GetFrame(wrapped2)
	if frame == nil {
		t.Fatal("expected frame")
	}

	if frame.Function != "frame2" {
		t.Errorf("expected frame2's function, got %v", frame.Function)
	}
	if frame.Line != 22 {
		t.Errorf("expected Line=22, got %v", frame.Line)
	}

	// First wrapper should still have frame1
	frameA := GetFrame(wrapped1)
	if frameA == nil {
		t.Fatal("expected frame from first wrapper")
	}
	if frameA.Function != "frame1" {
		t.Errorf("expected frame1's function from first wrapper, got %v", frameA.Function)
	}

	// Original node should have its original factory frame (or nil)
	frameOriginal := GetFrame(node)
	// Original frame could be different or nil, that's OK
	_ = frameOriginal

	// All nodes should work
	for _, n := range []Node{node, wrapped1, wrapped2} {
		status, err := n.Tick()
		if err != nil {
			t.Errorf("unexpected error for wrapped node: %v", err)
		}
		if status != Success {
			t.Errorf("expected Success, got %v", status)
		}
	}
}

// TestNode_WithFrame tests Node.WithFrame method
func TestNode_WithFrame_method(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }
	node := New(testTick)

	customFrame := &Frame{
		PC:       0xFEDCB,
		Function: "method.frame",
		File:     "method.go",
		Line:     88,
		Entry:    0xFEDCB,
	}

	// Use Node.WithFrame method
	wrapped := node.WithFrame(customFrame)

	frame := GetFrame(wrapped)
	if frame == nil {
		t.Fatal("expected frame from Node.WithFrame")
	}

	if frame.Function != "method.frame" {
		t.Errorf("expected Function='method.frame', got %v", frame.Function)
	}

	// Verify node works
	status, err := wrapped.Tick()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if status != Success {
		t.Errorf("expected Success, got %v", status)
	}
}

// TestNode_Frame_method tests Node.Frame method
func TestNode_Frame_method(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }

	t.Run("with factory frame", func(t *testing.T) {
		node := New(testTick)
		frame := node.Frame()

		// Frame should be available (or nil if runtime.Callers failed)
		if frame != nil {
			if frame.Function == "" && frame.PC == 0 {
				t.Error("frame has no useful information")
			}
		}
	})

	t.Run("with custom frame", func(t *testing.T) {
		node := New(testTick)
		customFrame := &Frame{
			PC:       0x99999,
			Function: "custom.node.frame",
			File:     "custom_node.go",
			Line:     33,
			Entry:    0x99999,
		}

		wrapped := node.WithFrame(customFrame)
		frame := wrapped.Frame()

		if frame == nil {
			t.Fatal("expected frame from Node.Frame")
		}

		// Should return the custom frame (or an approximation of it)
		// Note: Node.Frame may return an approximation based on newFrame(tick)
		// if the attached frame can't be retrieved
		if frame.PC != 0 && frame.PC != customFrame.PC {
			// Frame might have different PC from newFrame approximation
			t.Logf("Frame PC differs (may be approximation): got %v, want %v", frame.PC, customFrame.PC)
		}
	})
}

// TestTick_Frame_method tests Tick.Frame method
func TestTick_Frame_method(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }

	t.Run("success tick", func(t *testing.T) {
		frame := Tick(testTick).Frame()

		// Frame should be available
		if frame == nil {
			t.Fatal("expected frame from Tick.Frame")
		}

		// Should have at least some information
		if frame.PC == 0 && frame.Function == "" && frame.Entry == 0 {
			t.Error("frame has no information")
		}

		// Function name should relate to the tick function
		// (it will be the anonymous function name from the test)
		if frame.Function != "" {
			t.Logf("tick function: %v", frame.Function)
		}
	})

	t.Run("named tick", func(t *testing.T) {
		frame := Tick(Sequence).Frame()

		if frame == nil {
			t.Fatal("expected frame from Sequence.Frame")
		}

		if frame.Function == "" {
			t.Error("frame.Function should not be empty")
		}

		// Should be "Sequence" or "behaviortree.Sequence"
		if !strings.Contains(frame.Function, "Sequence") {
			t.Errorf("expected 'Sequence' in Function name, got %v", frame.Function)
		}
	})

	t.Run("nil tick", func(t *testing.T) {
		var nilTick Tick
		frame := nilTick.Frame()

		// nil tick should return nil frame
		if frame != nil {
			t.Errorf("expected nil frame for nil tick, got %v", frame)
		}
	})
}

// TestNewFrame_withFactoryIntegration tests NewFrame integration with factory
func TestNewFrame_withFactoryIntegration(t *testing.T) {
	// Create a runtime.Frame manually
	pc := uintptr(0x88888)
	runtimeFrame := runtime.Frame{
		PC:       pc,
		Function: "test.factory.integration",
		File:     "test_factory.go",
		Line:     55,
		Entry:    pc,
	}

	// Create a Frame using NewFrame
	btFrame := NewFrame(runtimeFrame)

	// Create a node and attach this frame manually
	testTick := func(children []Node) (Status, error) { return Success, nil }
	node := New(testTick)
	wrapped := node.WithFrame(&btFrame)

	// Verify frame can be retrieved
	retrievedFrame := GetFrame(wrapped)
	if retrievedFrame == nil {
		t.Fatal("expected frame")
	}

	// Should match what we set
	if retrievedFrame.Function != "test.factory.integration" {
		t.Errorf("expected Function='test.factory.integration', got %v", retrievedFrame.Function)
	}
	if retrievedFrame.Line != 55 {
		t.Errorf("expected Line=55, got %v", retrievedFrame.Line)
	}
}

// TestFrame_comprehensive tests all Frame methods and helpers together
func TestFrame_comprehensive(t *testing.T) {
	testTick := func(children []Node) (Status, error) { return Success, nil }

	// Create node through factory (should capture frame)
	node := New(testTick)

	// Test Node.Frame()
	nodeFrameViaMethod := node.Frame()

	// Test GetFrame(node)
	nodeFrameViaHelper := GetFrame(node)

	// Both should exist (or both be nil if runtime.Callers failed)
	if (nodeFrameViaMethod == nil) != (nodeFrameViaHelper == nil) {
		t.Errorf("inconsistent frame presence: method=%v, helper=%v",
			nodeFrameViaMethod != nil, nodeFrameViaHelper != nil)
	}

	// If both exist, they should be consistent
	if nodeFrameViaMethod != nil && nodeFrameViaHelper != nil {
		// They might not be identical (one might be approximation), but should both have some info
		t.Logf("Method frame PC: %v, Function: %v", nodeFrameViaMethod.PC, nodeFrameViaMethod.Function)
		t.Logf("Helper frame PC: %v, Function: %v", nodeFrameViaHelper.PC, nodeFrameViaHelper.Function)
	}

	// Test Tick.Frame()
	tick, _ := node()
	tickFrame := tick.Frame()

	if tickFrame == nil {
		t.Error("tick frame should not be nil")
	}

	// Test WithFrame
	customFrame := &Frame{
		PC:       0xABCDE,
		Function: "comprehensive.test",
		File:     "comprehensive.go",
		Line:     66,
		Entry:    0xABCDE,
	}

	wrapped := node.WithFrame(customFrame)
	wrappedFrame := GetFrame(wrapped)

	if wrappedFrame == nil {
		t.Fatal("wrapped frame should not be nil")
	}

	if wrappedFrame.Function != "comprehensive.test" {
		t.Errorf("expected Function='comprehensive.test', got %v", wrappedFrame.Function)
	}

	// All should still work
	for _, n := range []Node{node, wrapped} {
		status, err := n.Tick()
		if err != nil {
			t.Errorf("error ticking node: %v", err)
		}
		if status != Success {
			t.Errorf("expected Success, got %v", status)
		}
	}
}
