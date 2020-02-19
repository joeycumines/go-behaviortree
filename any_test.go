/*
   Copyright 2020 Joseph Cumines

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

func ExampleAny_sequencePartialSuccess() {
	fmt.Println(New(
		Any(Sequence),
		New(func(children []Node) (Status, error) {
			fmt.Println(1)
			return Success, nil
		}),
		New(func(children []Node) (Status, error) {
			fmt.Println(2)
			return Success, nil
		}),
		New(func(children []Node) (Status, error) {
			fmt.Println(3)
			return Success, nil
		}),
		New(func(children []Node) (Status, error) {
			fmt.Println(4)
			return Success, nil
		}),
		New(func(children []Node) (Status, error) {
			fmt.Println(5)
			return Failure, nil
		}),
		New(func(children []Node) (Status, error) {
			panic(`wont reach here`)
		}),
	).Tick())
	//output:
	//1
	//2
	//3
	//4
	//5
	//success <nil>
}

func ExampleAny_allPartialSuccess() {
	fmt.Println(New(
		Any(All),
		New(func(children []Node) (Status, error) {
			fmt.Println(1)
			return Success, nil
		}),
		New(func(children []Node) (Status, error) {
			fmt.Println(2)
			return Success, nil
		}),
		New(func(children []Node) (Status, error) {
			fmt.Println(3)
			return Success, nil
		}),
		New(func(children []Node) (Status, error) {
			fmt.Println(4)
			return Success, nil
		}),
		New(func(children []Node) (Status, error) {
			fmt.Println(5)
			return Failure, nil
		}),
		New(func(children []Node) (Status, error) {
			fmt.Println(6)
			return Success, nil
		}),
	).Tick())
	//output:
	//1
	//2
	//3
	//4
	//5
	//6
	//success <nil>
}

func ExampleAny_resetBehavior() {
	var (
		status Status
		err    error
		node   = New(
			Any(Sequence),
			New(func(children []Node) (Status, error) {
				fmt.Println(1)
				return status, err
			}),
			New(func(children []Node) (Status, error) {
				fmt.Println(2)
				return Success, nil
			}),
		)
	)

	status = Success
	err = nil
	fmt.Println(node.Tick())

	status = Failure
	err = nil
	fmt.Println(node.Tick())

	status = Success
	err = errors.New(`some_error`)
	fmt.Println(node.Tick())

	status = Success
	err = nil
	fmt.Println(node.Tick())

	//output:
	//1
	//2
	//success <nil>
	//1
	//failure <nil>
	//1
	//failure some_error
	//1
	//2
	//success <nil>
}

func ExampleAny_running() {
	status := Running
	node := New(
		Any(All),
		New(func(children []Node) (Status, error) {
			fmt.Printf("child ticked: %s\n", status)
			return status, nil
		}),
	)
	fmt.Println(node.Tick())
	status = Failure
	fmt.Println(node.Tick())
	status = Running
	fmt.Println(node.Tick())
	status = Success
	fmt.Println(node.Tick())
	//output:
	//child ticked: running
	//running <nil>
	//child ticked: failure
	//failure <nil>
	//child ticked: running
	//running <nil>
	//child ticked: success
	//success <nil>
}

func ExampleAny_forkPartialSuccess() {
	var (
		c1     = make(chan struct{})
		c2     = make(chan struct{})
		c3     = make(chan struct{})
		c4     = make(chan struct{})
		c5     = make(chan struct{})
		c6     = make(chan struct{})
		status = Running
	)
	go func() {
		time.Sleep(time.Millisecond * 100)
		fmt.Println(`unblocking the forked nodes`)
		close(c1)
		time.Sleep(time.Millisecond * 100)
		close(c2)
		time.Sleep(time.Millisecond * 100)
		close(c3)
		time.Sleep(time.Millisecond * 100)
		close(c4)
		time.Sleep(time.Millisecond * 100)
		close(c5)
		time.Sleep(time.Millisecond * 100)
		close(c6)
	}()
	node := New(
		Any(Fork()),
		New(func(children []Node) (Status, error) {
			fmt.Println(`ready`)
			<-c1
			fmt.Println(1)
			return Success, nil
		}),
		New(func(children []Node) (Status, error) {
			fmt.Println(`ready`)
			<-c2
			fmt.Println(2)
			return Success, nil
		}),
		New(func(children []Node) (Status, error) {
			fmt.Println(`ready`)
			<-c3
			fmt.Println(3)
			return status, nil
		}),
		New(func(children []Node) (Status, error) {
			fmt.Println(`ready`)
			<-c4
			fmt.Println(4)
			return Failure, nil
		}),
		New(func(children []Node) (Status, error) {
			fmt.Println(`ready`)
			<-c5
			fmt.Println(5)
			return Failure, nil
		}),
		New(func(children []Node) (Status, error) {
			fmt.Println(`ready`)
			<-c6
			fmt.Println(6)
			return Success, nil
		}),
	)
	fmt.Println(node.Tick())
	fmt.Println(`same running behavior as Fork`)
	fmt.Println(node.Tick())
	fmt.Println(`but the exit status is overridden`)
	status = Failure
	fmt.Println(node.Tick())
	//output:
	//ready
	//ready
	//ready
	//ready
	//ready
	//ready
	//unblocking the forked nodes
	//1
	//2
	//3
	//4
	//5
	//6
	//running <nil>
	//same running behavior as Fork
	//ready
	//3
	//running <nil>
	//but the exit status is overridden
	//ready
	//3
	//success <nil>
}

func TestAny_nilTick(t *testing.T) {
	if v := Any(nil); v != nil {
		t.Error(v)
	}
}

func TestAny_nilNode(t *testing.T) {
	if v, err := Any(Sequence)([]Node{nil}); err == nil || v != Failure {
		t.Error(v, err)
	}
}

func TestAny_nilChildTick(t *testing.T) {
	status, err := New(Any(Sequence), New(nil)).Tick()
	if status != Failure {
		t.Error(status)
	}
	if err == nil || err.Error() != `behaviortree.Node cannot tick a node with a nil tick` {
		t.Error(err)
	}
}

func TestAny_noChildren(t *testing.T) {
	if status, err := New(Any(Sequence)).Tick(); err != nil || status != Failure {
		t.Error(status, err)
	}
}
