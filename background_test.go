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

func ExampleBackground_success() {
	defer checkNumGoroutines(nil)(false, 0)
	node := func() Node {
		tick := Background(Fork)
		return func() (Tick, []Node) {
			return tick, []Node{
				New(func(children []Node) (Status, error) {
					fmt.Println(`start fork`)
					return Success, nil
				}),
				New(Async(func(children []Node) (Status, error) {
					time.Sleep(time.Millisecond * 100)
					return Success, nil
				})),
				New(Async(func(children []Node) (Status, error) {
					time.Sleep(time.Millisecond * 200)
					return Success, nil
				})),
				New(Async(func(children []Node) (Status, error) {
					time.Sleep(time.Millisecond * 300)
					fmt.Println(`end fork`)
					return Success, nil
				})),
			}
		}
	}()
	fmt.Println(node.Tick())
	time.Sleep(time.Millisecond * 50)
	fmt.Println(node.Tick())
	time.Sleep(time.Millisecond * 150)
	fmt.Println(node.Tick())
	time.Sleep(time.Millisecond * 200)
	fmt.Println(node.Tick()) // will receive the first tick's status
	time.Sleep(time.Millisecond * 50)
	fmt.Println(node.Tick())
	time.Sleep(time.Millisecond * 100)
	fmt.Println(node.Tick())
	fmt.Println(node.Tick())
	fmt.Println(node.Tick())
	time.Sleep(time.Millisecond * 450)
	fmt.Println(node.Tick())
	fmt.Println(node.Tick())
	//output:
	//start fork
	//running <nil>
	//start fork
	//running <nil>
	//start fork
	//running <nil>
	//end fork
	//end fork
	//success <nil>
	//success <nil>
	//end fork
	//success <nil>
	//start fork
	//running <nil>
	//start fork
	//running <nil>
	//end fork
	//end fork
	//success <nil>
	//success <nil>
}

func TestBackground(t *testing.T) {
	defer checkNumGoroutines(t)(false, 0)

	var (
		status Status
		err    error
		se     = errors.New(`some_error`)
	)
	node := New(Background(Fork), New(func(children []Node) (Status, error) {
		return status, err
	}))

	status, err = Success, nil
	if v, e := node.Tick(); v != Success || e != nil {
		t.Error(v, e)
	}

	status, err = Failure, nil
	if v, e := node.Tick(); v != Failure || e != nil {
		t.Error(v, e)
	}

	status, err = Success, se
	if v, e := node.Tick(); v != Failure || e != se {
		t.Error(v, e)
	}
}

func TestBackground_nilTick(t *testing.T) {
	if v := Background(nil); v != nil {
		t.Error(v)
	}
}

func TestBackground_withAny(t *testing.T) {
	defer checkNumGoroutines(t)(false, 0)

	var (
		tick = Background(func() Tick {
			return Any(All)
		})
	)

	one := make(chan Status)
	go func() {
		one <- Running
	}()
	if status, err := tick([]Node{
		New(func(children []Node) (status Status, e error) {
			return Failure, nil
		}),
		New(All, New(func(children []Node) (Status, error) {
			return <-one, nil
		})),
		New(func(children []Node) (status Status, e error) {
			return Failure, nil
		}),
	}); err != nil || status != Running {
		t.Error(status, err)
	}
	time.Sleep(time.Millisecond * 100)
	select {
	case <-one:
		t.Fatal(`expected consumed`)
	default:
	}

	two := make(chan Status)
	go func() {
		one <- Running
		two <- Running
	}()
	if status, err := tick([]Node{
		New(func(children []Node) (status Status, e error) {
			return Failure, nil
		}),
		New(All, New(func(children []Node) (Status, error) {
			return <-two, nil
		})),
		New(func(children []Node) (status Status, e error) {
			return Failure, nil
		}),
	}); err != nil || status != Running {
		t.Error(status, err)
	}
	time.Sleep(time.Millisecond * 100)
	select {
	case <-one:
		t.Fatal(`expected consumed`)
	case <-two:
		t.Fatal(`expected consumed`)
	default:
	}

	three := make(chan Status)
	go func() {
		one <- Running
		two <- Running
		three <- Running
	}()
	if status, err := tick([]Node{
		New(func(children []Node) (status Status, e error) {
			return Failure, nil
		}),
		New(All, New(func(children []Node) (Status, error) {
			return <-three, nil
		})),
		New(func(children []Node) (status Status, e error) {
			return Failure, nil
		}),
	}); err != nil || status != Running {
		t.Error(status, err)
	}
	time.Sleep(time.Millisecond * 100)
	select {
	case <-one:
		t.Fatal(`expected consumed`)
	case <-two:
		t.Fatal(`expected consumed`)
	case <-three:
		t.Fatal(`expected consumed`)
	default:
	}

	go func() {
		one <- Running
		two <- Running
		three <- Failure
	}()
	if status, err := tick([]Node{func() (tick Tick, nodes []Node) {
		panic(`should never reach here`)
	}}); err != nil || status != Failure {
		t.Error(status, err)
	}
	close(three)
	time.Sleep(time.Millisecond * 100)

	go func() {
		one <- Running
		two <- Running
	}()
	if status, err := tick([]Node{New(Selector)}); err != nil || status != Failure {
		t.Error(status, err)
	}
	time.Sleep(time.Millisecond * 100)
	select {
	case <-one:
		t.Fatal(`expected consumed`)
	case <-two:
		t.Fatal(`expected consumed`)
	default:
	}

	go func() {
		one <- Success
	}()
	if status, err := tick([]Node{func() (tick Tick, nodes []Node) {
		panic(`should never reach here`)
	}}); err != nil || status != Success {
		t.Error(status, err)
	}
	close(one)
	time.Sleep(time.Millisecond * 100)

	go func() {
		two <- Success
	}()
	if status, err := tick([]Node{func() (tick Tick, nodes []Node) {
		panic(`should never reach here`)
	}}); err != nil || status != Success {
		t.Error(status, err)
	}
	close(two)
	time.Sleep(time.Millisecond * 100)

	if status, err := tick([]Node{New(Sequence)}); err != nil || status != Success {
		t.Error(status, err)
	}
}
