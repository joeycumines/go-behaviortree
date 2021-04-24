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

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func ExampleShuffle() {
	rand.Seed(1231244)
	var (
		newPrintlnFn = func(fn func() []interface{}) Tick {
			return func([]Node) (Status, error) {
				fmt.Println(fn()...)
				return Success, nil
			}
		}
		newPrintln = func(v ...interface{}) Tick { return newPrintlnFn(func() []interface{} { return v }) }
		done       bool
		ticker     = NewTickerStopOnFailure(context.Background(), time.Millisecond, New(
			Sequence,
			New(newPrintlnFn(func() func() []interface{} {
				var i int
				return func() []interface{} {
					i++
					return []interface{}{`tick number`, i}
				}
			}())),
			New(
				Shuffle(Sequence, nil),
				New(newPrintln(`node 1`)),
				New(newPrintln(`node 2`)),
				New(
					Selector,
					New(func() func(children []Node) (Status, error) {
						remaining := 5
						return func(children []Node) (Status, error) {
							if remaining > 0 {
								remaining--
								return Success, nil
							}
							return Failure, nil
						}
					}()),
					New(
						Shuffle(Selector, nil),
						New(newPrintln(`node 3`)),
						New(newPrintln(`node 4`)),
						New(newPrintln(`node 5`)),
						New(newPrintln(`node 6`)),
						New(func([]Node) (Status, error) {
							done = true
							return Success, nil
						}),
					),
				),
			),
			New(func([]Node) (Status, error) {
				if done {
					return Failure, nil
				}
				return Success, nil
			}),
		))
	)
	<-ticker.Done()
	if err := ticker.Err(); err != nil {
		panic(err)
	}
	//output:
	//tick number 1
	//node 1
	//node 2
	//tick number 2
	//node 2
	//node 1
	//tick number 3
	//node 1
	//node 2
	//tick number 4
	//node 2
	//node 1
	//tick number 5
	//node 2
	//node 1
	//tick number 6
	//node 1
	//node 2
	//node 5
	//tick number 7
	//node 6
	//node 1
	//node 2
	//tick number 8
	//node 2
	//node 5
	//node 1
	//tick number 9
	//node 3
	//node 2
	//node 1
	//tick number 10
	//node 2
	//node 1
}

func TestShuffle_nil(t *testing.T) {
	if v := Shuffle(nil, nil); v != nil {
		t.Fatal(v)
	}
}
