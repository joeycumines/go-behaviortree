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

// All implements a tick which will tick all children sequentially until the first running status or error is
// encountered (propagated), and will return success only if all children were ticked and returned success (returns
// success if there were no children, like sequence).
func All(children []Node) (Status, error) {
	success := true
	for _, child := range children {
		status, err := child.Tick()
		if err != nil {
			return Failure, err
		}
		if status == Running {
			return Running, nil
		}
		if status != Success {
			success = false
		}
	}
	if !success {
		return Failure, nil
	}
	return Success, nil
}
