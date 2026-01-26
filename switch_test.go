/*
   Copyright 2026 Joseph Cumines

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
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestSwitch(t *testing.T) {
	type (
		ChildTick struct {
			Index  int
			Status Status
			Err    error
		}
		Invocation struct {
			Name       string
			Children   int
			ChildTicks []ChildTick
			Status     Status
			Err        error
		}
	)
	for _, tc := range [...]struct {
		Name        string
		Tick        Tick
		Invocations []Invocation
	}{
		{
			Name: `stateless`,
			Tick: Switch,
			Invocations: []Invocation{
				{
					Name:   `no children`,
					Status: Success,
				},
				{
					Name:     `default case success`,
					Children: 1,
					ChildTicks: []ChildTick{
						{0, Success, nil},
					},
					Status: Success,
				},
				{
					Name:     `default case invalid status and error`,
					Children: 1,
					ChildTicks: []ChildTick{
						{0, 123, errors.New(`some_error`)},
					},
					Status: 123,
					Err:    errors.New(`some_error`),
				},
				{
					Name:     `single case success`,
					Children: 2,
					ChildTicks: []ChildTick{
						{0, Success, nil},
						{1, Success, nil},
					},
					Status: Success,
				},
				{
					Name:     `single case failure`,
					Children: 2,
					ChildTicks: []ChildTick{
						{0, Success, nil},
						{1, Failure, nil},
					},
					Status: Failure,
				},
				{
					Name:     `single case success no match`,
					Children: 2,
					ChildTicks: []ChildTick{
						{0, Failure, nil},
					},
					Status: Success,
				},
				{
					Name:     `single case invalid status and error`,
					Children: 2,
					ChildTicks: []ChildTick{
						{0, Success, nil},
						{1, 123, errors.New(`some_error`)},
					},
					Status: 123,
					Err:    errors.New(`some_error`),
				},
				{
					Name:     `single case condition invalid status and error`,
					Children: 2,
					ChildTicks: []ChildTick{
						{0, 123, errors.New(`some_error`)},
					},
					Status: Failure,
					Err:    errors.New(`some_error`),
				},
				{
					Name:     `single case condition running`,
					Children: 2,
					ChildTicks: []ChildTick{
						{0, Running, nil},
					},
					Status: Running,
				},
				{
					Name:     `multi case`,
					Children: 9,
					ChildTicks: []ChildTick{
						{0, Failure, nil},
						{2, Failure, nil},
						{4, Success, nil},
						{5, Success, nil},
					},
					Status: Success,
				},
				{
					Name:     `multi case default failure`,
					Children: 7,
					ChildTicks: []ChildTick{
						{0, Failure, nil},
						{2, Failure, nil},
						{4, Failure, nil},
						{6, Failure, nil},
					},
					Status: Failure,
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			for _, invocation := range tc.Invocations {
				t.Run(invocation.Name, func(t *testing.T) {
					var children []Node
					for i := 0; i < invocation.Children; i++ {
						i := i
						children = append(children, New(func([]Node) (status Status, err error) {
							if len(invocation.ChildTicks) == 0 {
								t.Errorf(`child %d ticked but none expected`, i)
								return Success, nil
							}
							if invocation.ChildTicks[0].Index != i {
								t.Errorf(`child %d ticked but expected %d`, i, invocation.ChildTicks[0].Index)
								return Success, nil
							}
							status, err = invocation.ChildTicks[0].Status, invocation.ChildTicks[0].Err
							invocation.ChildTicks = invocation.ChildTicks[1:]
							return
						}))
					}
					status, err := tc.Tick(children)
					if (err == nil) != (invocation.Err == nil) || (err != nil && err.Error() != invocation.Err.Error()) {
						t.Error(err)
					}
					if status != invocation.Status {
						t.Error(status)
					}
					if len(invocation.ChildTicks) != 0 {
						t.Errorf(`expected %d more child ticks`, len(invocation.ChildTicks))
					}
				})
			}
		})
	}
}

func ExampleSwitch() {
	var (
		sanityChecks []func()
		newNode      = func(name string, statuses ...Status) Node {
			sanityChecks = append(sanityChecks, func() {
				if len(statuses) != 0 {
					panic(fmt.Errorf(`node %s has %d unconsumed statuses`, name, len(statuses)))
				}
			})
			return New(func([]Node) (status Status, _ error) {
				if len(statuses) == 0 {
					panic(fmt.Errorf(`node %s has no unconsumed statuses`, name))
				}
				status = statuses[0]
				statuses = statuses[1:]
				fmt.Printf("Tick %s: %s\n", name, status)
				return
			})
		}
		ticker = NewTickerStopOnFailure(
			context.Background(),
			time.Millisecond,
			New(
				Memorize(Sequence),
				newNode(`START`, Success, Success, Success, Success, Failure),
				New(
					Memorize(Selector),
					New(
						Memorize(Sequence),
						New(
							Memorize(Switch),

							newNode(`case-1-condition`, Failure, Failure, Running, Running, Running, Failure, Failure),
							newNode(`case-1-statement`),

							newNode(`case-2-condition`, Failure, Failure, Running, Running, Success, Success),
							newNode(`case-2-statement`, Running, Running, Running, Failure, Running, Success),

							newNode(`case-3-condition`, Failure, Failure),
							newNode(`case-3-statement`),

							newNode(`default-statement`, Failure, Success),
						),
						newNode(`SUCCESS`, Success, Success),
					),
					newNode(`FAILURE`, Success, Success),
				),
			),
		)
	)
	<-ticker.Done()
	if err := ticker.Err(); err != nil {
		panic(err)
	}
	for _, sanityCheck := range sanityChecks {
		sanityCheck()
	}
	// output:
	// Tick START: success
	// Tick case-1-condition: failure
	// Tick case-2-condition: failure
	// Tick case-3-condition: failure
	// Tick default-statement: failure
	// Tick FAILURE: success
	// Tick START: success
	// Tick case-1-condition: failure
	// Tick case-2-condition: failure
	// Tick case-3-condition: failure
	// Tick default-statement: success
	// Tick SUCCESS: success
	// Tick START: success
	// Tick case-1-condition: running
	// Tick case-1-condition: running
	// Tick case-1-condition: running
	// Tick case-1-condition: failure
	// Tick case-2-condition: running
	// Tick case-2-condition: running
	// Tick case-2-condition: success
	// Tick case-2-statement: running
	// Tick case-2-statement: running
	// Tick case-2-statement: running
	// Tick case-2-statement: failure
	// Tick FAILURE: success
	// Tick START: success
	// Tick case-1-condition: failure
	// Tick case-2-condition: success
	// Tick case-2-statement: running
	// Tick case-2-statement: success
	// Tick SUCCESS: success
	// Tick START: failure
}
