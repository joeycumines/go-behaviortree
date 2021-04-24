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

// Memorize encapsulates a tick, and will cache the first non-running status for each child, per "execution", defined
// as the period until the first non-running status, of the encapsulated tick, facilitating execution of asynchronous
// nodes in serial with their siblings, using stateless tick implementations, such as sequence and selector.
//
// Sync provides a similar but more flexible mechanism, at the expense of greater complexity, and more cumbersome
// usage. Sync supports modification of children mid-execution, and may be used to implement complex guarding behavior
// as children of a single Tick, equivalent to more complex structures using multiple memorized sequence nodes.
func Memorize(tick Tick) Tick {
	if tick == nil {
		return nil
	}
	var (
		started bool
		nodes   []Node
	)
	return func(children []Node) (status Status, err error) {
		if !started {
			nodes = copyNodes(children)
			for i := range nodes {
				var (
					child    = nodes[i]
					override Tick
				)
				if child == nil {
					continue
				}
				nodes[i] = func() (Tick, []Node) {
					tick, nodes := child()
					if override != nil {
						return override, nodes
					}
					if tick == nil {
						return nil, nodes
					}
					return func(children []Node) (Status, error) {
						status, err := tick(children)
						if err != nil || status != Running {
							override = func(children []Node) (Status, error) { return status, err }
						}
						return status, err
					}, nodes
				}
			}
			started = true
		}
		status, err = tick(nodes)
		if err != nil || status != Running {
			started = false
			nodes = nil
		}
		return
	}
}
