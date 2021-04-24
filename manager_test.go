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
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	startGoroutines := runtime.NumGoroutine()
	defer func() {
		time.Sleep(time.Millisecond * 100)
		endGoroutines := runtime.NumGoroutine()
		if startGoroutines < endGoroutines {
			t.Errorf("ended with %d more goroutines", endGoroutines-startGoroutines)
		}
	}()

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

	if len(m.tickers) != 2 || len(m.errs) != 0 || m.stopped {
		t.Fatal(m)
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

	if len(m.tickers) != 0 {
		t.Error(m.tickers)
	}

	if err := m.Err(); err == nil || err.Error() != "some_error | other_error" {
		t.Error(err)
	}

	// does nothing
	m.Stop()

	if err := m.Add(mockTicker{}); err == nil {
		t.Error("expected error")
	}
	m.errs = nil
	if err := m.Add(mockTicker{}); err == nil {
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
