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

import "time"

// RateLimit generates a stateful Tick that will return success at most once per a given duration
func RateLimit(d time.Duration) Tick {
	var last *time.Time
	return func(children []Node) (Status, error) {
		now := time.Now()
		if last != nil && now.Add(-d).Before(*last) {
			return Failure, nil
		}
		last = &now
		return Success, nil
	}
}
