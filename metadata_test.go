package behaviortree

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestNode_Metadata(t *testing.T) {
	t.Run("WithName", func(t *testing.T) {
		n := NewNode(func(children []Node) (Status, error) { return Success, nil }, nil)
		if n.Name() != "" {
			t.Error("expected empty name")
		}
		n2 := n.WithName("test")
		if n2.Name() != "test" {
			t.Errorf("expected name 'test', got '%s'", n2.Name())
		}
		if n.Name() != "" {
			t.Error("original node modified")
		}
	})

	t.Run("WithStructure", func(t *testing.T) {
		n := NewNode(func(children []Node) (Status, error) { return Success, nil }, nil)
		if n.Structure() != nil {
			t.Error("expected nil structure")
		}

		child := NewNode(func(children []Node) (Status, error) { return Success, nil }, nil)
		n2 := n.WithStructure(slices.Values([]Metadata{child}))

		s := slices.Collect(n2.Structure())
		if len(s) != 1 || fmt.Sprintf("%p", s[0]) != fmt.Sprintf("%p", child) {
			t.Errorf("structure mismatch")
		}

		if n.Structure() != nil {
			t.Error("original node modified")
		}
	})

	t.Run("WithStructure Clear", func(t *testing.T) {
		n := NewNode(func(children []Node) (Status, error) { return Success, nil }, nil)
		// Attach structure first
		n2 := n.WithStructure(slices.Values([]Metadata{n}))
		if n2.Structure() == nil {
			t.Error("expected non-nil structure")
		}
		// Clear it
		n3 := n2.WithStructure(nil)
		if n3.Structure() != nil {
			t.Error("expected nil structure")
		}
	})

	t.Run("WithStructure Interface Nil", func(t *testing.T) {
		n := NewNode(func(children []Node) (Status, error) { return Success, nil }, nil)
		n2 := n.WithStructure(nil)

		// Verify via Value() that it is raw interface nil, not typed nil
		val := n2.Value(vkStructure{})
		if val != nil {
			t.Errorf("expected nil interface, got %T: %v", val, val)
		}

		// Also verify Structure() works as expected
		if n2.Structure() != nil {
			t.Error("expected Structure() to return nil")
		}
	})

	t.Run("WithStructure Explicit Empty", func(t *testing.T) {
		n := NewNode(func(children []Node) (Status, error) { return Success, nil }, nil)
		n2 := n.WithStructure(func(yield func(Metadata) bool) {})
		seq := n2.Structure()
		if seq == nil {
			t.Error("expected non-nil sequence")
		}
		s := slices.Collect(seq)
		if len(s) != 0 {
			t.Error("expected empty slice")
		}
	})
}

func TestWalk(t *testing.T) {
	// Construct a tree where physical structure differs from logical structure
	// Root -> (Physical: ChildA)
	// Root.Structure -> (ChildB)

	var visited []string

	childA := NewNode(func(children []Node) (Status, error) { return Success, nil }, nil).WithName("ChildA")
	childB := NewNode(func(children []Node) (Status, error) { return Success, nil }, nil).WithName("ChildB")

	root := NewNode(func(children []Node) (Status, error) { return Success, nil }, []Node{childA}).WithName("Root")
	rootWithStructure := root.WithStructure(slices.Values([]Metadata{childB}))

	Walk(rootWithStructure, func(n Metadata) bool {
		if node, ok := n.(Node); ok {
			visited = append(visited, node.Name())
		}
		return true
	})

	// Expect: Root, ChildB. (ChildA should be skipped because Structure overrides)
	expected := []string{"Root", "ChildB"}
	if !reflect.DeepEqual(visited, expected) {
		t.Errorf("expected %v, got %v", expected, visited)
	}

	// Test Clearing (Exposing underlying children)
	visited = nil
	// rootWithStructure has ChildB. passing nil reverts to root's behavior (ChildA)
	//
	// NOTE: This creates outer: rootRestored -> inner: rootWithStructure -> root
	// When rootRestored.WithStructure(nil) is called, it creates a wrapper that returns nil
	// for Structure(). However, physical children come from the underlying chain:
	// rootRestored() -> rootWithStructure() -> root() which still returns [childA].
	// The Walker checks Structure() first, gets nil, then calls n() to get physical children.
	// Physical children reflect the actual execution path through the chain, which yields childA.
	// This is correct behavior: WithStructure(nil) clears the metadata, exposing the physical structure.
	rootRestored := rootWithStructure.WithStructure(nil)
	Walk(rootRestored, func(n Metadata) bool {
		if node, ok := n.(Node); ok {
			visited = append(visited, node.Name())
		}
		return true
	})
	expected = []string{"Root", "ChildA"}
	if !reflect.DeepEqual(visited, expected) {
		t.Errorf("expected %v, got %v", expected, visited)
	}

	// Test Explicit Masking
	visited = nil
	rootMasked := root.WithStructure(func(yield func(Metadata) bool) {})
	Walk(rootMasked, func(n Metadata) bool {
		if node, ok := n.(Node); ok {
			visited = append(visited, node.Name())
		}
		return true
	})
	expected = []string{"Root"}
	if !reflect.DeepEqual(visited, expected) {
		t.Errorf("expected %v, got %v", expected, visited)
	}

	// Test Default expansion in Walk
	visited = nil
	Walk(root, func(n Metadata) bool {
		if node, ok := n.(Node); ok {
			visited = append(visited, node.Name())
		}
		return true
	})
	expected = []string{"Root", "ChildA"}
	if !reflect.DeepEqual(visited, expected) {
		t.Errorf("expected %v, got %v", expected, visited)
	}
}

func TestNode_String_Name(t *testing.T) {
	n := NewNode(func(children []Node) (Status, error) { return Success, nil }, nil).WithName("MyNode")
	s := n.String()
	// Output should look something like: [meta...]  MyNode
	if !strings.Contains(s, "MyNode") {
		t.Errorf("expected string to contain 'MyNode', got:\n%s", s)
	}
}
