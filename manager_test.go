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
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestManager_Stop_raceCloseDone(t *testing.T) {
	defer checkNumGoroutines(t)(false, 0)
	m := NewManager().(*manager)
	close(m.done)
	m.Stop()
}

func TestManager_Stop_noTickers(t *testing.T) {
	defer checkNumGoroutines(t)(false, 0)
	m := NewManager()
	if err := m.Err(); err != nil {
		t.Error(err)
	}
	select {
	case <-m.Done():
		t.Error()
	default:
	}
	m.Stop()
	if err := m.Err(); err != nil {
		t.Error(err)
	}
	<-m.Done()
	if err := m.Err(); err != nil {
		t.Error(err)
	}
}

func TestManager_Add_whileStopping(t *testing.T) {
	defer checkNumGoroutines(t)(false, 0)
	m := NewManager()
	for i := 0; i < 10; i++ {
		if err := m.Add(NewManager()); err != nil {
			t.Fatal(err)
		}
	}
	var (
		wg      sync.WaitGroup
		count   int64
		done    int32
		stopped int32
	)
	wg.Add(8)
	defer func() {
		atomic.AddInt64(&count, -atomic.LoadInt64(&count))
		m.Stop()
		atomic.StoreInt32(&stopped, 1)
		time.Sleep(time.Millisecond * 50)
		atomic.StoreInt32(&done, 1)
		wg.Wait()
		<-m.Done()
		if err := m.Err(); err != nil {
			t.Error(err)
		}
		count := atomic.LoadInt64(&count)
		t.Log(count)
		if count < 15 {
			t.Error(count)
		}
	}()
	for i := 0; i < 8; i++ {
		go func() {
			defer wg.Done()
			for atomic.LoadInt32(&done) == 0 {
				var (
					stoppedBefore = atomic.LoadInt32(&stopped) != 0
					err           = m.Add(NewManager())
					stoppedAfter  = atomic.LoadInt32(&stopped) != 0
				)
				if err != nil && err != ErrManagerStopped {
					t.Error(err)
				}
				if stoppedBefore && !stoppedAfter {
					t.Error()
				}
				if stoppedBefore && err == nil {
					t.Error()
				}
				atomic.AddInt64(&count, 1)
				time.Sleep(time.Millisecond)
			}
		}()
	}
	time.Sleep(time.Millisecond * 20)
}

func TestManager_Add_secondStopCase(t *testing.T) {
	defer checkNumGoroutines(t)(false, 0)
	out := make(chan error)
	defer close(out)
	done := make(chan struct{})
	stop := make(chan struct{})
	go func() { out <- (&manager{stop: stop, done: done}).Add(mockTicker{}) }()
	time.Sleep(time.Millisecond * 100)
	select {
	case err := <-out:
		t.Fatal(err)
	default:
	}
	close(stop)
	if err := <-out; err != ErrManagerStopped {
		t.Error(err)
	}
	<-done
}

func TestManager_Stop_cleanupGoroutines(t *testing.T) {
	check := checkNumGoroutines(t)
	defer check(false, 0)

	m := NewManager()

	{
		// add a ticker then stop it, then verify that all resources (goroutines) are cleaned up
		check(false, 0)
		done := make(chan struct{})
		if err := m.Add(mockTicker{
			done: func() <-chan struct{} { return done },
			err:  func() error { return nil },
			stop: func() {},
		}); err != nil {
			t.Fatal(err)
		}
		check(true, 0)
		close(done)
		check(false, 0)
		if err := m.Err(); err != nil {
			t.Error(err)
		}
		select {
		case <-m.Done():
			t.Fatal()
		default:
		}
	}

	{
		// add two tickers (one multiple times), then stop one, then the other
		var (
			m1 = NewManager()
			m2 = NewManager()
		)
		check(false, 0)
		for i := 0; i < 30; i++ {
			if err := m.Add(m1); err != nil {
				t.Fatal(err)
			}
		}
		if err := m.Add(m2); err != nil {
			t.Fatal(err)
		}
		check(true, 0)
		gr := runtime.NumGoroutine()
		m1.Stop()
		if diff := waitNumGoroutines(0, func(n int) bool { return (n - gr + 1) <= 0 }) - gr + 1; diff > 0 {
			t.Errorf("too many goroutines: +%d", diff)
		}
		m2.Stop()
	}
}

func TestNewManager(t *testing.T) {
	defer checkNumGoroutines(t)(false, 0)

	m := NewManager().(*manager)

	select {
	case <-m.Done():
		t.Error("unexpected done")
	default:
	}

	if err := m.Err(); err != nil {
		t.Fatal(err)
	}

	var (
		mutex  sync.Mutex
		stops1 int
		done1  = make(chan struct{})
		err1   error
		stops2 int
		done2  = make(chan struct{})
		err2   error
	)

	if err := m.Add(
		mockTicker{
			done: func() <-chan struct{} {
				return done1
			},
			err: func() error {
				mutex.Lock()
				defer mutex.Unlock()
				return err1
			},
			stop: func() {
				mutex.Lock()
				defer mutex.Unlock()
				stops1++
			},
		},
	); err != nil {
		t.Fatal(err)
	}

	if err := m.Add(
		mockTicker{
			done: func() <-chan struct{} {
				return done2
			},
			err: func() error {
				mutex.Lock()
				defer mutex.Unlock()
				return err2
			},
			stop: func() {
				mutex.Lock()
				defer mutex.Unlock()
				stops2++
			},
		},
	); err != nil {
		t.Fatal(err)
	}

	if d := m.Done(); d != m.done || d == nil {
		t.Error(d)
	}
	if err := m.Err(); err != nil {
		t.Error(err)
	}
	select {
	case <-m.stop:
		t.Error()
	default:
	}
	select {
	case <-m.done:
		t.Error()
	default:
	}

	mutex.Lock()
	err2 = errors.New("some_error")
	close(done2)
	mutex.Unlock()

	time.Sleep(time.Millisecond * 100)

	mutex.Lock()
	if stops2 != 1 {
		t.Error(stops2)
	}
	if stops1 != 1 {
		t.Error(stops1)
	}
	err1 = errors.New("other_error")
	mutex.Unlock()

	select {
	case <-m.Done():
		t.Error("unexpected done")
	default:
	}

	close(done1)

	<-m.Done()

	checkErrTicker := func(err error) {
		t.Helper()
		if err == nil || err.Error() != "some_error | other_error" {
			t.Error(err)
		}
		if !errors.Is(err, err1) {
			t.Error(err)
		}
		if !errors.Is(err, err2) {
			t.Error(err)
		}
		if errors.Is(err, errors.New(`another_error`)) {
			t.Error(err)
		}
		{
			err := err
			if v, ok := err.(errManagerStopped); ok {
				err = v.Unwrap()
			}
			if v, ok := err.(errManagerTicker); !ok || len(v) != 2 {
				t.Error(err)
			}
		}
	}
	checkErrStopped := func(err error) {
		t.Helper()
		if !errors.Is(err.(errManagerStopped), ErrManagerStopped.(errManagerStopped)) ||
			!(errManagerStopped{}).Is(err) ||
			!err.(interface{ Is(error) bool }).Is(errManagerStopped{}) {
			t.Error(err)
		}
	}

	checkErrTicker(m.Err())
	{
		err := m.Add(mockTicker{})
		checkErrTicker(err)
		checkErrStopped(err)
	}

	// does nothing
	m.Stop()

	checkErrTicker(m.Err())
	{
		err := m.Add(mockTicker{})
		checkErrTicker(err)
		checkErrStopped(err)
	}

	m.errs = nil
	if err := m.Add(mockTicker{}); err != ErrManagerStopped {
		t.Error("expected error")
	}
	if err := m.Add(nil); err == nil {
		t.Error("expected error")
	}
}

type mockTicker struct {
	done func() <-chan struct{}
	err  func() error
	stop func()
}

func (m mockTicker) Done() <-chan struct{} {
	if m.done != nil {
		return m.done()
	}
	panic("implement me")
}

func (m mockTicker) Err() error {
	if m.err != nil {
		return m.err()
	}
	panic("implement me")
}

func (m mockTicker) Stop() {
	if m.stop != nil {
		m.stop()
		return
	}
	panic("implement me")
}

const (
	waitNumGoroutinesDefault     = time.Millisecond * 200
	waitNumGoroutinesNumerator   = 1
	waitNumGoroutinesDenominator = 1
	waitNumGoroutinesMin         = time.Millisecond * 50
)

func waitNumGoroutines(wait time.Duration, fn func(n int) bool) (n int) {
	if wait == 0 {
		wait = waitNumGoroutinesDefault
	}
	wait *= waitNumGoroutinesNumerator
	wait /= waitNumGoroutinesDenominator
	if wait < waitNumGoroutinesMin {
		wait = waitNumGoroutinesMin
	}
	count := int(wait / waitNumGoroutinesMin)
	wait /= time.Duration(count)
	n = runtime.NumGoroutine()
	for i := 0; i < count && !fn(n); i++ {
		time.Sleep(wait)
		runtime.GC()
		n = runtime.NumGoroutine()
	}
	return
}

// checkNumGoroutines is used to indirectly test goroutine state / cleanup, the reliance on timing isn't great, but I
// wasn't able to come up with a better solution
func checkNumGoroutines(t *testing.T) func(increase bool, wait time.Duration) {
	if t != nil {
		t.Helper()
	}
	var (
		errorf = func(format string, values ...interface{}) {
			if err := fmt.Errorf(format, values...); t != nil {
				t.Error(err)
			} else {
				panic(err)
			}
		}
		start = runtime.NumGoroutine()
	)
	return func(increase bool, wait time.Duration) {
		if t != nil {
			t.Helper()
		}
		var fn func(n int) bool
		if increase {
			fn = func(n int) bool { return start < n }
		} else {
			fn = func(n int) bool { return start >= n }
		}
		if now := waitNumGoroutines(wait, fn); increase {
			if start >= now {
				errorf("too few goroutines: -%d", start-now+1)
			}
		} else if start < now {
			errorf("too many goroutines: +%d", now-start)
		}
	}
}
