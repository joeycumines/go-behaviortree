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

// Package behaviortree provides a simple and powerful Go implementation of behavior trees without fluff.
package behaviortree

import (
	"errors"
	"fmt"
)

const (
	_ = iota
	// Running indicates that the Tick for a given Node is currently running
	Running Status = iota
	// Success indicates that the Tick for a given Node completed successfully
	Success
	// Failure indicates that the Tick for a given Node failed to complete successfully
	Failure
)

type (
	// Node represents an node in a tree, that can be ticked
	Node func() (Tick, []Node)

	// Tick represents the logic for a node, which may or may not be stateful
	Tick func(children []Node) (Status, error)

	// Status is a type with three valid values, Running, Success, and Failure, the three possible states for BTs
	Status int
)

// New constructs a new behavior tree aliasing NewNode with vararg support for less indentation
func New(tick Tick, children ...Node) Node {
	return NewNode(tick, children)
}

// NewNode constructs a new node out of a tick and children
func NewNode(tick Tick, children []Node) Node {
	return func() (Tick, []Node) {
		return tick, children
	}
}

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

// Status returns the status value, but defaults to failure on out of bounds
func (s Status) Status() Status {
	switch s {
	case Running:
		return Running
	case Success:
		return Success
	default:
		return Failure
	}
}

// String returns a string representation of the status
func (s Status) String() string {
	switch s {
	case Running:
		return `running`
	case Success:
		return `success`
	case Failure:
		return `failure`
	default:
		return fmt.Sprintf("unknown status (%d)", s)
	}
}
