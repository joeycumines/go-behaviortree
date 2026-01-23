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

// vkName is the context key for Node.Name
type vkName struct{}

// vkStructure is the context key for Node.Structure
type vkStructure struct{}

// WithName returns a copy of the receiver, wrapped with the name value attached, for access via Node.Name.
func (n Node) WithName(name string) Node {
	return n.WithValue(vkName{}, name)
}

// Name returns the name value of the node, or an empty string.
func (n Node) Name() string {
	if n == nil {
		return ""
	}
	v, _ := n.Value(vkName{}).(string)
	return v
}

// WithStructure returns a copy of the receiver, wrapped with the structure value attached, for access via Node.Structure.
//
// Structure should be used to provide the "logical" children of a node, for cases where the actual children (attached
// via closure) do not accurately represent the node's semantics (e.g. FSM acting as a leaf, or a decorator), or where
// the node is a leaf but has internal structure relevant to inspection.
//
// Passing no children (or nil children) will cause Node.Structure to return a non-nil empty slice, which indicates
// that the node has structure but it is empty (effectively masking any real children from the walker).
func (n Node) WithStructure(children ...Node) Node {
	if children == nil {
		children = make([]Node, 0)
	}
	return n.WithValue(vkStructure{}, children)
}

// Structure returns the structure value of the node, or nil.
//
// A nil return indicates that no structure value was attached (and typically the walker should fall back to expansion).
// A non-nil empty slice indicates that the structure is explicitly empty.
func (n Node) Structure() []Node {
	if n == nil {
		return nil
	}
	v, _ := n.Value(vkStructure{}).([]Node)
	return v
}

// Walk will traverse the tree depth-first, preferring Structure() over actual expansion if present.
func Walk(n Node, fn func(n Node)) {
	if n == nil {
		return
	}
	fn(n)
	if s := n.Structure(); s != nil {
		for _, child := range s {
			Walk(child, fn)
		}
	} else {
		_, children := n()
		for _, child := range children {
			Walk(child, fn)
		}
	}
}
