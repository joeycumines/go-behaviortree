/*
   Copyright 2018 Joseph Cumines

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

import "fmt"

// Sequence is a tick implementation that will succeed if any children succeed, returning running if any children
// return running, propagating any error
func Sequence(children []Node) (Status, error) {
	for i, c := range children {
		status, err := c.Tick()
		if err != nil {
			return Failure, fmt.Errorf("bt.Sequence encountered error with child at index %d: %s", i, err.Error())
		}
		if status == Running {
			return Running, nil
		}
		if status == Failure {
			return Failure, nil
		}
	}
	return Success, nil
}
