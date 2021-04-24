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

import "sync"

// Sync will wrap a set of nodes in such a way that their real ticks will only be triggered when either the node
// being ticked was previously running, or no other nodes are running, synchronising calling their Node and Tick calls.
//
// NOTE the Memorize function provides similar functionality, and should be preferred, where both are suitable.
func Sync(nodes []Node) []Node {
	if nodes == nil {
		return nil
	}
	s := &syc{
		nodes:    nodes,
		mutex:    new(sync.Mutex),
		statuses: make([]Status, len(nodes)),
	}
	result := make([]Node, 0, len(nodes))
	for i := range nodes {
		result = append(result, s.node(i))
	}
	return result
}

type syc struct {
	nodes    []Node
	statuses []Status
	mutex    *sync.Mutex
}

func (s *syc) running() bool {
	for _, status := range s.statuses {
		if status == Running {
			return true
		}
	}
	return false
}

func (s *syc) node(i int) Node {
	if s.nodes[i] == nil {
		return nil
	}
	return func() (Tick, []Node) {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		tick, children := s.nodes[i]()
		if tick == nil {
			return nil, children
		}
		status := s.statuses[i]
		if status != Running && s.running() {
			return func(children []Node) (Status, error) {
				// disabled tick - we just return the last status
				return status, nil
			}, children
		}
		return func(children []Node) (Status, error) {
			s.mutex.Lock()
			defer s.mutex.Unlock()
			status, err := tick(children)
			s.statuses[i] = status
			return status, err
		}, children
	}
}
