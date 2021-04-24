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

// Not inverts a Tick, such that any failure cases will be success and success cases will be failure, note that any
// error or invalid status will still result in a failure
func Not(tick Tick) Tick {
	if tick == nil {
		return nil
	}
	return func(children []Node) (Status, error) {
		status, err := tick(children)
		if err != nil {
			return Failure, err
		}
		switch status {
		case Running:
			return Running, nil
		case Failure:
			return Success, nil
		default:
			return Failure, nil
		}
	}
}
