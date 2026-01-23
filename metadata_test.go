package behaviortree

import (
	"fmt"
	"reflect"
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
		n2 := n.WithStructure(child)

		s := n2.Structure()
		if len(s) != 1 || fmt.Sprintf("%p", s[0]) != fmt.Sprintf("%p", child) {
			t.Errorf("structure mismatch")
		}

		if n.Structure() != nil {
			t.Error("original node modified")
		}
	})

	t.Run("WithStructure Masking", func(t *testing.T) {
		n := NewNode(func(children []Node) (Status, error) { return Success, nil }, nil)
		n2 := n.WithStructure() // Empty
		s := n2.Structure()
		if s == nil {
			t.Error("expected non-nil empty slice")
		}
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
	rootWithStructure := root.WithStructure(childB)

	Walk(rootWithStructure, func(n Node) {
		visited = append(visited, n.Name())
	})

	// Expect: Root, ChildB. (ChildA should be skipped because Structure overrides)
	expected := []string{"Root", "ChildB"}
	if !reflect.DeepEqual(visited, expected) {
		t.Errorf("expected %v, got %v", expected, visited)
	}

	// Test Masking
	visited = nil
	rootMasked := root.WithStructure()
	Walk(rootMasked, func(n Node) {
		visited = append(visited, n.Name())
	})
	expected = []string{"Root"}
	if !reflect.DeepEqual(visited, expected) {
		t.Errorf("expected %v, got %v", expected, visited)
	}

	// Test Default expansion in Walk
	visited = nil
	Walk(root, func(n Node) {
		visited = append(visited, n.Name())
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
