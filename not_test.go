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
