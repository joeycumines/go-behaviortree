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
	"context"
	"fmt"
	"time"
)

// ExampleNewTickerStopOnFailure_counter demonstrates the use of NewTickerStopOnFailure to implement more complex "run
// to completion" behavior using the simple modular building blocks provided by this package
func ExampleNewTickerStopOnFailure_counter() {
	var (
		// counter is the shared state used by this example
		counter = 0
		// printCounter returns a node that will print the counter prefixed with the given name then succeed
		printCounter = func(name string) Node {
			return New(
				func(children []Node) (Status, error) {
					fmt.Printf("%s: %d\n", name, counter)
					return Success, nil
				},
			)
		}
		// incrementCounter is a node that will increment counter then succeed
		incrementCounter = New(
			func(children []Node) (Status, error) {
				counter++
				return Success, nil
			},
		)
		// ticker is what actually runs this example and will tick the behavior tree defined by a given node at a given
		// rate and will stop after the first failed tick or error or context cancel
		ticker = NewTickerStopOnFailure(
			context.Background(),
			time.Millisecond,
			New(
				Selector, // runs each child sequentially until one succeeds (success) or all fail (failure)
				New(
					Sequence, // runs each child in order until one fails (failure) or they all succeed (success)
					New(
						func(children []Node) (Status, error) { // succeeds while counter is less than 10
							if counter < 10 {
								return Success, nil
							}
							return Failure, nil
						},
					),
					incrementCounter,
					printCounter("< 10"),
				),
				New(
					Sequence,
					New(
						func(children []Node) (Status, error) { // succeeds while counter is less than 20
							if counter < 20 {
								return Success, nil
							}
							return Failure, nil
						},
					),
					incrementCounter,
					printCounter("< 20"),
				),
			), // if both children failed (counter is >= 20) the root node will also fail
		)
	)
	// waits until ticker stops, which will be on the first failure of it's root node
	<-ticker.Done()
	// every Tick may return an error which would automatically cause a failure and propagation of the error
	if err := ticker.Err(); err != nil {
		panic(err)
	}
	// Output:
	// < 10: 1
	// < 10: 2
	// < 10: 3
	// < 10: 4
	// < 10: 5
	// < 10: 6
	// < 10: 7
	// < 10: 8
	// < 10: 9
	// < 10: 10
	// < 20: 11
	// < 20: 12
	// < 20: 13
	// < 20: 14
	// < 20: 15
	// < 20: 16
	// < 20: 17
	// < 20: 18
	// < 20: 19
	// < 20: 20
}
