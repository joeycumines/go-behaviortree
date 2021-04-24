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
	"testing"
)

func TestNot(t *testing.T) {
	children := make([]Node, 3)
	if status, err := Not(func(children []Node) (Status, error) {
		if len(children) != 3 {
			t.Error(children)
		}
		return Failure, errors.New(`some_err`)
	})(children); status != Failure || err == nil || err.Error() != `some_err` {
		t.Error(status, err)
	}
	if status, err := Not(func(children []Node) (Status, error) {
		if len(children) != 3 {
			t.Error(children)
		}
		return Running, nil
	})(children); status != Running || err != nil {
		t.Error(status, err)
	}
	if status, err := Not(func(children []Node) (Status, error) {
		if len(children) != 3 {
			t.Error(children)
		}
		return Failure, nil
	})(children); status != Success || err != nil {
		t.Error(status, err)
	}
	if status, err := Not(func(children []Node) (Status, error) {
		if len(children) != 3 {
			t.Error(children)
		}
		return Success, nil
	})(children); status != Failure || err != nil {
		t.Error(status, err)
	}
	if status, err := Not(func(children []Node) (Status, error) {
		if len(children) != 3 {
			t.Error(children)
		}
		return 1243145, nil
	})(children); status != Failure || err != nil {
		t.Error(status, err)
	}
	if Not(nil) != nil {
		t.Fatal(`wat`)
	}
}
