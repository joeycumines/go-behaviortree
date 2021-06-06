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
	"fmt"
	"testing"
	"time"
)

func TestFork_success(t *testing.T) {
	defer checkNumGoroutines(t)(false, 0)

	var (
		out    = make(chan string)
		prefix = `1.`
		node   = New(
			Fork(),
			New(Async(func(children []Node) (Status, error) {
				out <- prefix + `1`
				return Success, nil
			})),
			New(Async(func(children []Node) (Status, error) {
				out <- prefix + `2`
				return Success, nil
			})),
			New(Async(func(children []Node) (Status, error) {
				out <- prefix + `3`
				return Success, nil
			})),
		)
		seen = make(map[string]struct{})
		tick = func() Status {
			select {
			case value := <-out:
				if _, ok := seen[value]; ok {
					t.Fatal(value)
				}
				seen[value] = struct{}{}
			default:
			}
			time.Sleep(time.Millisecond * 100)
			status, err := node.Tick()
			if err != nil {
				t.Fatal(err)
			}
			return status
		}
	)

	for x := 1; x <= 3; x++ {
		seen = make(map[string]struct{})
		prefix = fmt.Sprintf("%d.", x)
		time.Sleep(time.Millisecond * 100)
		select {
		case value := <-out:
			t.Fatal(value)
		default:
		}
		if status, err := node.Tick(); err != nil || status != Running {
			t.Fatal(status, err)
		}
		time.Sleep(time.Millisecond * 100)
		if status := tick(); status != Running {
			t.Fatal(status, seen)
		}
		if status := tick(); status != Running {
			t.Fatal(status, seen)
		}
		if status := tick(); status != Success {
			t.Fatal(status, seen)
		}
		if len(seen) != 3 {
			t.Fatal(seen)
		}
		for y := 1; y <= 3; y++ {
			if _, ok := seen[fmt.Sprintf("%d.%d", x, y)]; !ok {
				t.Fatal(seen)
			}
		}
	}
}

func TestFork_failure(t *testing.T) {
	defer checkNumGoroutines(t)(false, 0)

	var (
		out    = make(chan string)
		prefix = `1.`
		node   = New(
			Fork(),
			New(Async(func(children []Node) (Status, error) {
				out <- prefix + `1`
				return Success, nil
			})),
			New(Async(func(children []Node) (Status, error) {
				out <- prefix + `2`
				return Success, nil
			})),
			New(func(children []Node) (Status, error) {
				return 0, errors.New(`error_one`)
			}),
			New(Async(func(children []Node) (Status, error) {
				out <- prefix + `3`
				return Success, errors.New(`error_two`)
			})),
		)
		seen = make(map[string]struct{})
		tick = func() Status {
			select {
			case value := <-out:
				if _, ok := seen[value]; ok {
					t.Fatal(value)
				}
				seen[value] = struct{}{}
			default:
			}
			time.Sleep(time.Millisecond * 100)
			status, err := node.Tick()
			if status == Running {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				if err == nil || err.Error() != `error_one | error_two` {
					t.Fatal(err)
				}
			}
			return status
		}
	)

	for x := 0; x < 3; x++ {
		seen = make(map[string]struct{})
		time.Sleep(time.Millisecond * 100)
		select {
		case value := <-out:
			t.Fatal(value)
		default:
		}
		if status, err := node.Tick(); err != nil || status != Running {
			t.Fatal(status, err)
		}
		time.Sleep(time.Millisecond * 100)
		if status := tick(); status != Running {
			t.Fatal(status, seen)
		}
		if status := tick(); status != Running {
			t.Fatal(status, seen)
		}
		if status := tick(); status != Failure {
			t.Fatal(status, seen)
		}
	}
}

func TestFork_noChildren(t *testing.T) {
	node := New(Fork())
	if status, err := node.Tick(); status != Success || err != nil {
		t.Fatal(status, err)
	}
	if status, err := node.Tick(); status != Success || err != nil {
		t.Fatal(status, err)
	}
	if status, err := node.Tick(); status != Success || err != nil {
		t.Fatal(status, err)
	}
}

func TestFork_immediateSuccess(t *testing.T) {
	node := New(
		Fork(),
		New(func(children []Node) (Status, error) {
			return Success, nil
		}),
		New(func(children []Node) (Status, error) {
			return Success, nil
		}),
		New(func(children []Node) (Status, error) {
			return Success, nil
		}),
	)
	if status, err := node.Tick(); status != Success || err != nil {
		t.Fatal(status, err)
	}
	if status, err := node.Tick(); status != Success || err != nil {
		t.Fatal(status, err)
	}
	if status, err := node.Tick(); status != Success || err != nil {
		t.Fatal(status, err)
	}
}

func TestFork_immediateFailure(t *testing.T) {
	node := New(
		Fork(),
		New(func(children []Node) (Status, error) {
			return Success, nil
		}),
		New(func(children []Node) (Status, error) {
			return 0, nil
		}),
		New(func(children []Node) (Status, error) {
			return Success, nil
		}),
	)
	if status, err := node.Tick(); status != Failure || err != nil {
		t.Fatal(status, err)
	}
	if status, err := node.Tick(); status != Failure || err != nil {
		t.Fatal(status, err)
	}
	if status, err := node.Tick(); status != Failure || err != nil {
		t.Fatal(status, err)
	}
}

func TestFork_immediateError(t *testing.T) {
	node := New(
		Fork(),
		New(func(children []Node) (Status, error) {
			return Success, errors.New(`err`)
		}),
		New(func(children []Node) (Status, error) {
			return Success, errors.New(`err`)
		}),
		New(func(children []Node) (Status, error) {
			return Success, errors.New(`err`)
		}),
	)
	if status, err := node.Tick(); status != Failure || err == nil || err.Error() != `err | err | err` {
		t.Fatal(status, err)
	}
	if status, err := node.Tick(); status != Failure || err == nil || err.Error() != `err | err | err` {
		t.Fatal(status, err)
	}
	if status, err := node.Tick(); status != Failure || err == nil || err.Error() != `err | err | err` {
		t.Fatal(status, err)
	}
}
