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

// Background pushes running nodes into the background, allowing multiple concurrent ticks (potentially running
// independent children, depending on the behavior of the node). It accepts a tick via closure, in order to support
// stateful ticks. On tick, backgrounded nodes are ticked from oldest to newest, until the first non-running status is
// returned, which will trigger removal from the backgrounded node list, and propagating status and any error, without
// modification. All other normal operation will result in a new node being generated and ticked, backgrounding it on
// running, otherwise discarding the node and propagating it's return values immediately. Passing a nil value will
// cause nil to be returned.
// WARNING there is no upper bound to the number of backgrounded nodes (the caller must manage that externally).
func Background(tick func() Tick) Tick {
	if tick == nil {
		return nil
	}
	var nodes []Node
	return func(children []Node) (Status, error) {
		for i, node := range nodes {
			status, err := node.Tick()
			if err == nil && status == Running {
				continue
			}
			copy(nodes[i:], nodes[i+1:])
			nodes[len(nodes)-1] = nil
			nodes = nodes[:len(nodes)-1]
			return status, err
		}
		node := NewNode(tick(), children)
		status, err := node.Tick()
		if err != nil || status != Running {
			return status, err
		}
		nodes = append(nodes, node)
		return Running, nil
	}
}
