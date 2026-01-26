package behaviortree

import (
	"slices"
	"testing"
)

// noOpNode is a simple leaf node
var noOpNode = NewNode(func(children []Node) (Status, error) {
	return Success, nil
}, nil)

// buildDeep constructs a linear chain of depth n
func buildDeep(n int) Node {
	current := noOpNode
	for i := 0; i < n; i++ {
		// Wrap in a parent
		child := current
		current = NewNode(func(children []Node) (Status, error) {
			return Success, nil
		}, []Node{child})
	}
	return current
}

// buildWide constructs a single root with n children
func buildWide(n int) Node {
	children := make([]Node, n)
	for i := 0; i < n; i++ {
		children[i] = noOpNode
	}
	return NewNode(func(c []Node) (Status, error) {
		return Success, nil
	}, children)
}

// buildStructureDeep constructs a linear chain where children are exposed via Structure()
func buildStructureDeep(n int) Node {
	current := noOpNode
	for i := 0; i < n; i++ {
		// Create a parent that HAS no physical children, but exposes 'current' via Metadata
		child := current
		parent := NewNode(func(children []Node) (Status, error) {
			return Success, nil
		}, nil)
		current = parent.WithStructure(slices.Values([]Metadata{child}))
	}
	return current
}

// buildFull constructs a full N-ary tree of height h
// Total nodes = (N^(h+1) - 1) / (N - 1)
// For N=5, h=4 (depth 5 levels): (5^5 - 1) / 4 = (3125 - 1)/4 = 781
func buildFull(branching, height int) Node {
	if height == 0 {
		return noOpNode
	}
	children := make([]Node, branching)
	for i := 0; i < branching; i++ {
		children[i] = buildFull(branching, height-1)
	}
	return NewNode(func(c []Node) (Status, error) {
		return Success, nil
	}, children)
}

func BenchmarkWalk_Deep100(b *testing.B) {
	root := buildDeep(100)
	count := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Walk(root, func(n Metadata) bool {
			count++
			return true
		})
	}
	b.StopTimer()
	if expected := b.N * 101; count != expected {
		b.Fatalf("expected %d nodes, got %d", expected, count)
	}
}

func BenchmarkWalk_Wide100(b *testing.B) {
	root := buildWide(100)
	count := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Walk(root, func(n Metadata) bool {
			count++
			return true
		})
	}
	b.StopTimer()
	if expected := b.N * 101; count != expected {
		b.Fatalf("expected %d nodes, got %d", expected, count)
	}
}

func BenchmarkWalk_LargeTree(b *testing.B) {
	// N=5, Height=4 => 5 levels (root, + 5, + 25, + 125, + 625) = 781 nodes
	root := buildFull(5, 4)
	count := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Walk(root, func(n Metadata) bool {
			count++
			return true
		})
	}
	b.StopTimer()
	if expected := b.N * 781; count != expected {
		b.Fatalf("expected %d nodes, got %d", expected, count)
	}
}

func BenchmarkWalk_StructureDeep100(b *testing.B) {
	root := buildStructureDeep(100)
	count := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Walk(root, func(n Metadata) bool {
			count++
			return true
		})
	}
	b.StopTimer()
	if expected := b.N * 101; count != expected {
		b.Fatalf("expected %d nodes, got %d", expected, count)
	}
}

// simpleMetadata is a custom Metadata implementation that avoids Node.Value overhead
type simpleMetadata struct {
	children []Metadata
}

func (s *simpleMetadata) Value(key any) any { return nil }

func (s *simpleMetadata) Children(yield func(Metadata) bool) {
	for _, c := range s.children {
		if !yield(c) {
			return
		}
	}
}

func BenchmarkWalk_StructureDeep100_Optimized(b *testing.B) {
	// Build a deep tree using custom metadata structs instead of Nodes
	var current Metadata = noOpNode
	for i := 0; i < 100; i++ {
		current = &simpleMetadata{children: []Metadata{current}}
	}

	count := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Walk(current, func(n Metadata) bool {
			count++
			return true
		})
	}
	b.StopTimer()
	if expected := b.N * 101; count != expected {
		b.Fatalf("expected %d nodes, got %d", expected, count)
	}
}
