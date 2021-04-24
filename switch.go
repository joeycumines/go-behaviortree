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

// Switch is a tick implementation that provides switch-like functionality, where each switch case is comprised of a
// condition and statement, formed by a pair of (contiguous) children. If there are an odd number of children, then the
// final child will be treated as a statement with an always-true condition (used as the default case). The first error
// or first running status will be returned (if any). Otherwise, the result will be either that of the statement
// corresponding to the first successful condition, or success.
//
// This implementation is compatible with both Memorize and Sync.
func Switch(children []Node) (Status, error) {
	for i := 0; i < len(children); i += 2 {
		if i == len(children)-1 {
			// statement (default case)
			return children[i].Tick()
		}
		// condition (normal case)
		status, err := children[i].Tick()
		if err != nil {
			return Failure, err
		}
		if status == Running {
			return Running, nil
		}
		if status == Success {
			// statement (normal case)
			return children[i+1].Tick()
		}
	}
	// no matching condition and no default statement
	return Success, nil
}
