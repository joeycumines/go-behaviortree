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

// Async wraps a tick so that it runs asynchronously, note nil ticks will return nil
func Async(tick Tick) Tick {
	if tick == nil {
		return nil
	}
	var done chan struct {
		Status Status
		Error  error
	}
	return func(children []Node) (Status, error) {
		if done == nil {
			// start the async tick, the non-nil done indicates that we are running
			done = make(chan struct {
				Status Status
				Error  error
			}, 1)
			go func() {
				var status struct {
					Status Status
					Error  error
				}
				defer func() {
					done <- status
				}()
				status.Status, status.Error = tick(children)
			}()
			return Running, nil
		}
		// the node is currently running
		select {
		case status := <-done:
			done = nil
			return status.Status, status.Error
		default:
			return Running, nil
		}
	}
}
