package behaviortree

import (
	"testing"
	"fmt"
	"context"
	"time"
	"reflect"
	"sync"
	"runtime"
	"errors"
)

func TestNewTicker_panic1(t *testing.T) {
	defer func() {
		r := recover()
		if s := fmt.Sprint(r); r == nil || s != "behaviortree.NewTicker nil context" {
			t.Fatal("unexpected panic", s)
		}
	}()
	NewTicker(nil, 1, func() (Tick, []Node) {
		return nil, nil
	})
	t.Error("expected a panic")
}

func TestNewTicker_panic2(t *testing.T) {
	defer func() {
		r := recover()
		if s := fmt.Sprint(r); r == nil || s != "behaviortree.NewTicker duration <= 0" {
			t.Fatal("unexpected panic", s)
		}
	}()
	NewTicker(context.Background(), 0, func() (Tick, []Node) {
		return nil, nil
	})
	t.Error("expected a panic")
}

func TestNewTicker_panic3(t *testing.T) {
	defer func() {
		r := recover()
		if s := fmt.Sprint(r); r == nil || s != "behaviortree.NewTicker nil node" {
			t.Fatal("unexpected panic", s)
		}
	}()
	NewTicker(context.Background(), 1, nil)
	t.Error("expected a panic")
}

func TestNewTicker_run(t *testing.T) {
	startGoroutines := runtime.NumGoroutine()
	defer func() {
		time.Sleep(time.Millisecond * 50)

		endGoroutines := runtime.NumGoroutine()

		if endGoroutines > startGoroutines {
			t.Error("started with", startGoroutines, "goroutines and ended with", endGoroutines)
		}
	}()

	var (
		mutex   sync.Mutex
		counter int
	)

	node := func() (Tick, []Node) {
		return func(children []Node) (Status, error) {
			mutex.Lock()
			defer mutex.Unlock()
			counter++
			time.Sleep(time.Millisecond)
			return Success, nil
		}, nil
	}

	c := NewTicker(
		context.WithValue(context.Background(), 1, 2),
		time.Millisecond*5,
		node,
	)

	time.Sleep(time.Millisecond * 50)

	v := c.(*nodeTicker)

	if v.node == nil || reflect.ValueOf(node).Pointer() != reflect.ValueOf(v.node).Pointer() {
		//noinspection GoPrintFunctions
		t.Error("unexpected node", v.node)
	}

	if v.ctx == nil || v.ctx.Value(1) != 2 {
		t.Error("unexpected context", v.ctx)
	}

	if v.cancel == nil {
		t.Error("nil cancel")
	}

	if v.err != nil {
		t.Error("unexpected error", v.err)
	}

	if v.stop == nil || v.done == nil {
		t.Error("nil chans")
	}

	if v.ticker == nil {
		t.Error("nil ticker")
	}

	func() {
		mutex.Lock()
		defer mutex.Unlock()
		if counter < 0 {
			t.Error("bad counter", counter)
		}
	}()

	if err := c.Err(); err != nil {
		t.Error("unexpected error call", err)
	}

	if d := c.Done(); d == nil || d != v.done {
		t.Error("unexpected done call", d)
	}

	func() {
		start := time.Now()

		c.Stop()

		<-c.Done()

		diff := time.Now().Sub(start)

		if diff > time.Millisecond*20 {
			t.Fatal("unexpected diff", diff)
		}

		if err := c.Err(); err != nil {
			t.Fatal(err)
		}
	}()
}

func TestNewTicker_runPanic(t *testing.T) {
	startGoroutines := runtime.NumGoroutine()
	defer func() {
		time.Sleep(time.Millisecond * 50)

		endGoroutines := runtime.NumGoroutine()

		if endGoroutines > startGoroutines {
			t.Error("started with", startGoroutines, "goroutines and ended with", endGoroutines)
		}
	}()

	node := func() (Tick, []Node) {
		panic("some_panic")
	}

	startedAt := time.Now()
	defer func() {
		diff := time.Now().Sub(startedAt)
		if diff > time.Millisecond*20 {
			t.Error("unexpected diff", diff)
		}
	}()

	c := NewTicker(
		context.Background(),
		time.Millisecond,
		node,
	)

	<-c.Done()

	if err := c.Err(); err == nil || err.Error() != "recovered from panic (string): some_panic" {
		t.Error("unexpected error", err)
	}
}

func TestNewTicker_runError(t *testing.T) {
	startGoroutines := runtime.NumGoroutine()
	defer func() {
		time.Sleep(time.Millisecond * 50)

		endGoroutines := runtime.NumGoroutine()

		if endGoroutines > startGoroutines {
			t.Error("started with", startGoroutines, "goroutines and ended with", endGoroutines)
		}
	}()

	expected := errors.New("some_error")

	node := func() (Tick, []Node) {
		return func(children []Node) (Status, error) {
			return 0, expected
		}, nil
	}

	startedAt := time.Now()
	defer func() {
		diff := time.Now().Sub(startedAt)
		if diff > time.Millisecond*20 {
			t.Error("unexpected diff", diff)
		}
	}()

	c := NewTicker(
		context.Background(),
		time.Millisecond,
		node,
	)

	<-c.Done()

	if err := c.Err(); err != expected {
		t.Error("unexpected error", err)
	}
}

func TestNewTicker_runCancel(t *testing.T) {
	startGoroutines := runtime.NumGoroutine()
	defer func() {
		time.Sleep(time.Millisecond * 50)

		endGoroutines := runtime.NumGoroutine()

		if endGoroutines > startGoroutines {
			t.Error("started with", startGoroutines, "goroutines and ended with", endGoroutines)
		}
	}()

	node := func() (Tick, []Node) {
		return func(children []Node) (Status, error) {
			time.Sleep(time.Millisecond)
			return Success, nil
		}, nil
	}

	startedAt := time.Now()
	defer func() {
		diff := time.Now().Sub(startedAt)
		if diff > time.Millisecond*20 {
			t.Error("unexpected diff", diff)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*5)
	defer cancel()

	c := NewTicker(
		ctx,
		time.Millisecond,
		node,
	)

	<-c.Done()

	if err := c.Err(); err == nil || err.Error() != "context deadline exceeded" {
		t.Error("unexpected error", err)
	}
}
