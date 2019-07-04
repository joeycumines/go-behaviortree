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

package behaviortree

import (
	"testing"
	"time"
)

func TestRateLimit(t *testing.T) {
	var (
		count    = 0
		duration = time.Millisecond * 100
		node     = New(
			Sequence,
			New(RateLimit(duration)),
			New(func(children []Node) (Status, error) {
				count++
				return Success, nil
			}),
		)
	)
	timer := time.NewTimer(duration*4 + (duration / 2))
	defer timer.Stop()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if _, err := node.Tick(); err != nil {
				t.Fatal(err)
			}
			continue
		case <-timer.C:
		}
		if count != 5 {
			t.Fatal(count)
		}
		return
	}
}
