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
	"errors"
	"github.com/go-test/deep"
	"testing"
	"time"
)

func TestAsync_behaviour(t *testing.T) {
	out := make(chan int, 50)
	assertOut := func(expected []int) {
		actual := make([]int, 0)
		for running := true; running; {
			select {
			case i := <-out:
				actual = append(actual, i)
			default:
				running = false
			}
		}
		if diff := deep.Equal(expected, actual); diff != nil {
			t.Fatal("unexpected out diff:", diff)
		}
	}

	node := NewNode(
		Sequence,
		Sync([]Node{
			NewNode(func(children []Node) (Status, error) {
				out <- 1
				return Success, nil
			}, nil),
			NewNode(
				Async(func(children []Node) (Status, error) {
					time.Sleep(time.Millisecond * 100)
					out <- 2
					return Success, nil
				}),
				nil,
			),
			NewNode(func(children []Node) (Status, error) {
				out <- 3
				return Success, nil
			}, nil),
		}),
	)

	run := func() {
		assertOut([]int{})
		if status, err := node.Tick(); status != Running {
			t.Error("status was meant to be running but it was", status)
		} else if err != nil {
			t.Error("error was non-nil", err)
		}
		assertOut([]int{1})
		// still running
		if status, err := node.Tick(); status != Running {
			t.Error("status was meant to be running but it was", status)
		} else if err != nil {
			t.Error("error was non-nil", err)
		}
		// sleep for a bit
		time.Sleep(time.Millisecond * 50)
		// still running
		assertOut([]int{})
		if status, err := node.Tick(); status != Running {
			t.Error("status was meant to be running but it was", status)
		} else if err != nil {
			t.Error("error was non-nil", err)
		}
		assertOut([]int{})
		// sleep for a bit more
		time.Sleep(time.Millisecond * 70)
		assertOut([]int{2})
		// should be done
		if status, err := node.Tick(); status != Success {
			t.Error("status was meant to be running but it was", status)
		} else if err != nil {
			t.Error("error was non-nil", err)
		}
		assertOut([]int{3})
	}

	for x := 0; x < 3; x++ {
		run()
	}
}

func TestAsync_error(t *testing.T) {
	rErr := errors.New("some_error")
	node := NewNode(
		Async(func(children []Node) (Status, error) {
			return Failure, rErr
		}),
		nil,
	)
	if status, err := node.Tick(); status != Running {
		t.Error("status was meant to be running but it was", status)
	} else if err != nil {
		t.Error("error was non-nil", err)
	}
	time.Sleep(time.Millisecond * 5)
	if status, err := node.Tick(); status != Failure {
		t.Error("status was meant to be failure but it was", status)
	} else if err != rErr {
		t.Error("unexpected error value:", err)
	}
}

func TestAsync_nil(t *testing.T) {
	if Async(nil) != nil {
		t.Fatal("expected nil tick")
	}
}
