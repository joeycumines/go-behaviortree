package behaviortree

import (
	"testing"
)

func BenchmarkNode_Tick_Baseline(b *testing.B) {
	node := NewNode(func(children []Node) (Status, error) { return Success, nil }, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node.Tick()
	}
}

func BenchmarkNode_Tick_Wrapped_NoValue(b *testing.B) {
	node := NewNode(func(children []Node) (Status, error) { return Success, nil }, nil)
	// Wrap it but don't query value in tick
	wrapped := node.WithName("bench")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wrapped.Tick()
	}
}
