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
	"errors"
	"testing"
)

func TestMemorize_nilTick(t *testing.T) {
	if v := Memorize(nil); v != nil {
		t.Error(`expected nil`)
	}
}

func TestMemorize_nilChildCases(t *testing.T) {
	var (
		i int
		j int
		e = errors.New(`some_error`)
	)
	_, _ = New(
		Memorize(func(children []Node) (status Status, err error) {
			if len(children) != 3 ||
				children[0] == nil ||
				children[1] != nil ||
				children[2] == nil {
				t.Fatal(children)
			}
			{
				tick, c := children[0]()
				if tick != nil {
					t.Error(`expected nil`)
				}
				if len(c) != 1 || c[0] == nil {
					t.Error(c)
				}
				if status, err := c[0].Tick(); err != nil || status != Failure {
					t.Error(status, err)
				}
			}
			for x := 0; x < 10; x++ {
				tick, c := children[2]()
				if c != nil {
					t.Error(c)
				}
				status, err := tick(c)
				if status != Running || err != e {
					t.Error(status, err)
				}
			}
			i++
			return
		}),
		New(nil, New(Selector)),
		nil,
		New(func(children []Node) (status Status, err error) {
			j++
			status = Running
			err = e
			return
		}),
	).Tick()
	if i != 1 {
		t.Error(i)
	}
	if j != 1 {
		t.Error(j)
	}
}

func TestMemorize_errorResets(t *testing.T) {
	var (
		i    int
		j    int
		e    error
		node = New(
			Memorize(func(children []Node) (Status, error) {
				// this is just sequence but without normalisation of status to failure on error
				for _, c := range children {
					status, err := c.Tick()
					if status != Success || err != nil {
						return status, err
					}
				}
				return Success, nil
			}),
			New(func([]Node) (Status, error) {
				i++
				return Success, nil
			}),
			New(func([]Node) (Status, error) {
				j++
				return Running, e
			}),
		)
	)
	if i != 0 || j != 0 {
		t.Fatal(i, j)
	}
	if status, err := node.Tick(); err != nil || status != Running {
		t.Fatal(status, err)
	}
	if i != 1 || j != 1 {
		t.Fatal(i, j)
	}
	if status, err := node.Tick(); err != nil || status != Running {
		t.Fatal(status, err)
	}
	if i != 1 || j != 2 {
		t.Fatal(i, j)
	}
	if status, err := node.Tick(); err != nil || status != Running {
		t.Fatal(status, err)
	}
	if i != 1 || j != 3 {
		t.Fatal(i, j)
	}
	e = errors.New(`some_error`)
	if status, err := node.Tick(); err != e || status != Running {
		t.Fatal(status, err)
	}
	if i != 1 || j != 4 {
		t.Fatal(i, j)
	}
	e = nil
	if status, err := node.Tick(); err != nil || status != Running {
		t.Fatal(status, err)
	}
	if i != 2 || j != 5 {
		t.Fatal(i, j)
	}
}
