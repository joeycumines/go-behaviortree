/*
   Copyright 2019 Joseph Cumines

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

// Package behaviortree provides a functional style implementation of behavior trees in Golang with no fluff
package behaviortree

import "errors"

type (
	// Node represents an node in a tree, that can be ticked
	Node func() (Tick, []Node)

	// Tick represents the logic for a node, which may or may not be stateful
	Tick func(children []Node) (Status, error)
)

// Tick runs the node's tick function with it's children
func (n Node) Tick() (Status, error) {
	if n == nil {
		return Failure, errors.New("behaviortree.Node cannot tick a nil node")
	}
	tick, children := n()
	if tick == nil {
		return Failure, errors.New("behaviortree.Node cannot tick a node with a nil tick")
	}
	return tick(children)
}

// NewNode constructs a new node out of a tick and children
func NewNode(tick Tick, children []Node) Node {
	return func() (Tick, []Node) {
		return tick, children
	}
}
