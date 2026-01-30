package behaviortree

import (
	"fmt"
	"iter"
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

	t.Run("WithName Clear", func(t *testing.T) {
		n := NewNode(func([]Node) (Status, error) { return Success, nil }, nil).WithName("foo")
		n2 := n.WithName("")
		if n2.Name() != "" {
			t.Error("Name should be empty")
		}
		// Verify Value returns nil (interface)
		if v := n2.Value(vkName{}); v != nil {
			t.Errorf("Value should be nil, got %v", v)
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

func TestUseName(t *testing.T) {
	// 1. Test direct provider usage
	name := "test-provider-name"
	p := UseName(name)

	// Check correct key
	if v, ok := p.Value(vkName{}); !ok || v != name {
		t.Errorf("Value(vkName{}) = %v, %v; want %v, true", v, ok, name)
	}

	// Check incorrect key
	if _, ok := p.Value("wrong"); ok {
		t.Error("Value(wrong key) returned true")
	}

	// 2. Integration with Node
	// This node simulates a factory that uses UseName
	var node Node = func() (Tick, []Node) {
		UseValueProvider(UseName("integrated-name"))
		return func(children []Node) (Status, error) { return Success, nil }, nil
	}

	if n := node.Name(); n != "integrated-name" {
		t.Errorf("Node.Name() = %q; want %q", n, "integrated-name")
	}

	// Verify using GetName helper
	if n := GetName(node); n != "integrated-name" {
		t.Errorf("GetName(node) = %q; want %q", n, "integrated-name")
	}

	// 3. Test Clearing (Empty String)
	pClear := UseName("")
	if v, ok := pClear.Value(vkName{}); !ok {
		t.Error("UseName(\"\").Value returned false")
	} else if v != nil {
		t.Errorf("UseName(\"\").Value = %v; want nil", v)
	}
}

func TestUseStructure(t *testing.T) {
	// 1. Test direct provider
	child := NewNode(func([]Node) (Status, error) { return Success, nil }, nil)
	seq := slices.Values([]Metadata{child})

	p := UseStructure(seq)

	// Check correct key
	if v, ok := p.Value(vkStructure{}); !ok {
		t.Error("Value(vkStructure{}) returned false")
	} else {
		// Verify exact value identity if possible, or behavior
		s, ok := v.(iter.Seq[Metadata])
		if !ok {
			t.Errorf("Value returned %T; want iter.Seq[Metadata]", v)
		} else {
			collected := slices.Collect(s)
			if len(collected) != 1 || fmt.Sprintf("%p", collected[0]) != fmt.Sprintf("%p", child) {
				t.Errorf("Structure content mismatch: %v", collected)
			}
		}
	}

	// Check incorrect key
	if _, ok := p.Value("wrong"); ok {
		t.Error("Value(wrong key) returned true")
	}

	// 2. Integration with Node
	var node Node = func() (Tick, []Node) {
		UseValueProvider(UseStructure(seq))
		return func([]Node) (Status, error) { return Success, nil }, nil
	}

	s := node.Structure()
	if s == nil {
		t.Fatal("Node.Structure() returned nil")
	}
	found := slices.Collect(s)
	if len(found) != 1 || fmt.Sprintf("%p", found[0]) != fmt.Sprintf("%p", child) {
		t.Errorf("Node.Structure() mismatch")
	}

	// 3. Test Nil
	nilP := UseStructure(nil)
	if v, ok := nilP.Value(vkStructure{}); !ok {
		t.Error("UseStructure(nil).Value returned false")
	} else if v != nil {
		t.Errorf("UseStructure(nil).Value = %v; want nil", v)
	}

	// 4. Test Explicit Empty
	emptySeq := func(yield func(Metadata) bool) {}
	pEmpty := UseStructure(emptySeq)
	if v, ok := pEmpty.Value(vkStructure{}); !ok {
		t.Error("UseStructure(empty).Value returned false")
	} else {
		if v == nil {
			t.Error("UseStructure(empty).Value is nil")
		} else if s, ok := v.(iter.Seq[Metadata]); !ok {
			t.Errorf("UseStructure(empty).Value type = %T; want iter.Seq[Metadata]", v)
		} else if len(slices.Collect(s)) != 0 {
			t.Error("UseStructure(empty) sequence is not empty")
		}
	}
}

func TestWalk_EarlyStop(t *testing.T) {
	child := NewNode(func(children []Node) (Status, error) { return Success, nil }, nil).WithName("Child")
	root := NewNode(func(children []Node) (Status, error) { return Success, nil }, []Node{child}).WithName("Root")

	// 1. Stop at Root
	stops := 0
	Walk(root, func(n Metadata) bool {
		stops++
		return false
	})
	if stops != 1 {
		t.Errorf("Walk didn't stop at root, visited %d nodes", stops)
	}

	// 2. Stop at Child
	visited := []string{}
	Walk(root, func(n Metadata) bool {
		if node, ok := n.(Node); ok {
			visited = append(visited, node.Name())
			if node.Name() == "Child" {
				return false
			}
		}
		return true
	})
	if len(visited) != 2 {
		t.Errorf("Walk didn't stop at child: %v", visited)
	}
}

func TestChildren_YieldStop(t *testing.T) {
	child1 := NewNode(func(children []Node) (Status, error) { return Success, nil }, nil).WithName("Child1")
	child2 := NewNode(func(children []Node) (Status, error) { return Success, nil }, nil).WithName("Child2")
	root := NewNode(func(children []Node) (Status, error) { return Success, nil }, []Node{child1, child2})

	count := 0
	root.Children(func(m Metadata) bool {
		count++
		return false
	})
	if count != 1 {
		t.Errorf("Children yield didn't stop, count=%d", count)
	}
}
