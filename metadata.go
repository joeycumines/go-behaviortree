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

import "iter"

// Metadata represents the "conceptual" structure of a behavior tree or subtree, which may or may not correspond to
// actual [Node] instances.
//
// This interface allows for efficient traversal and introspection of tree structures without necessarily incurring the
// overhead of [Node.Value] for every single node, effectively allowing whole subtrees to be "virtualized" or
// generated on demand.
//
// Note: [Node] implements this interface.
type Metadata interface {
	// Value returns the value associated with the given key, or nil if not present.
	// This loosely corresponds to [context.Context]'s `Value` method.
	Value(key any) any

	// Children yields the logical children of this metadata node.
	// Returning false from the yield function stops iteration.
	Children(yield func(Metadata) bool)
}

var _ Metadata = Node(nil)

// Valuer is an interface for types that can return values associated with keys.
//
// This interface along with [ValueAttachable] allows for generic handling of key-value attachments
// without tying to a specific implementation. It is intended for implementation of [context.Context]-style
// `With*` and `Get*` exported functions, used to keep the keys private while allowing external packages to
// attach and retrieve values.
//
// Note: [Node] implements this interface.
type Valuer interface {
	// Value returns the value associated with the given key, or nil if not present.
	// This loosely corresponds to [context.Context]'s `Value` method.
	Value(key any) any
}

var _ Valuer = Node(nil)

// ValueAttachable is an interface for types that support attaching key-value pairs.
//
// This interface along with [Valuer] allows for generic handling of key-value attachments
// without tying to a specific implementation. It is intended for implementation of [context.Context]-style
// `With*` and `Get*` exported functions, used to keep the keys private while allowing external packages to
// attach and retrieve values.
//
// Note: [Node] implements this interface.
type ValueAttachable[T any] interface {
	WithValue(key, value any) T
}

var _ ValueAttachable[Node] = Node(nil)

// vkName is the context key for Node.Name
type vkName struct{}

func GetName(n Valuer) string {
	v, _ := n.Value(vkName{}).(string)
	return v
}

func WithName[T any](n ValueAttachable[T], name string) T {
	return n.WithValue(vkName{}, name)
}

// vkStructure is the context key for Node.Structure
type vkStructure struct{}

// WithName returns a copy of the receiver, wrapped with the name value attached, for access via Node.Name.
func (n Node) WithName(name string) Node {
	return WithName[Node](n, name)
}

// Name returns the name value of the node, or an empty string.
func (n Node) Name() string {
	return GetName(n)
}

// WithStructure returns a copy of the receiver, wrapped with the structure value attached, for access via Node.Structure.
//
// Structure provides the "logical" children of a node, allowing the tree's conceptual structure to differ from its
// physical implementation (closures). This is useful for:
//   - Decorators or wrappers that should appear as a single node or transparent.
//   - Complex leaf nodes (like FSMs) that want to expose internal state as a subtree.
//   - Optimizing traversal by providing a [Metadata] sequence that avoids the [Node.Value] lock overhead for children.
//
// Passing a nil sequence will cause Node.Structure to return nil, clearing any previous structure and reverting to
// physical node expansion. To explicitly mask children (making the node appear as a leaf), pass an empty sequence:
// func(yield func(Metadata) bool) {}.
func (n Node) WithStructure(children iter.Seq[Metadata]) Node {
	if children == nil {
		return n.WithValue(vkStructure{}, nil)
	}
	return n.WithValue(vkStructure{}, children)
}

// Structure returns the structure value of the node, or nil.
//
// A nil return indicates that no structure value was attached (and typically the walker should fall back to expansion).
// A non-nil empty sequence indicates that the structure is explicitly empty.
func (n Node) Structure() iter.Seq[Metadata] {
	v, _ := n.Value(vkStructure{}).(iter.Seq[Metadata])
	return v
}

// Walk traverses the "conceptual" tree structure starting from n, depth-first.
//
// It uses the `Metadata` interface to determine children, preferring `n.Structure()` (logical children) over
// physical node expansion if present. This allows for rich, efficient introspection of complex or stateful trees.
func Walk(n Metadata, fn func(n Metadata) bool) {
	walk(n, fn)
}

func walk(n Metadata, fn func(n Metadata) bool) bool {
	if !fn(n) {
		return false
	}
	stopped := false
	n.Children(func(child Metadata) bool {
		if !walk(child, fn) {
			stopped = true
			return false
		}
		return true
	})
	return !stopped
}

// Children yields the logical children of the node.
//
// It prefers [Node.Structure] (logical children) over physical node expansion if present. This allows for rich,
// efficient introspection of complex or stateful trees.
func (n Node) Children(yield func(Metadata) bool) {
	if s := n.Structure(); s != nil {
		s(yield)
		return
	}

	_, children := n()
	for _, child := range children {
		if !yield(child) {
			return
		}
	}
}
