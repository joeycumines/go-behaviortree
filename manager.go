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
	"github.com/joeycumines/go-bigbuff"
	"sync"
)

type (
	// Manager models an aggregate Ticker, which should stop gracefully on the first failure
	Manager interface {
		Ticker

		// Add will register a new ticker under this manager
		Add(ticker Ticker) error
	}

	// manager is this package's implementation of the Manager interface
	manager struct {
		mu      sync.RWMutex
		once    sync.Once
		worker  bigbuff.Worker
		done    chan struct{}
		stop    chan struct{}
		tickers chan managerTicker
		errs    []error
	}

	managerTicker struct {
		Ticker Ticker
		Done   func()
	}

	errManagerTicker []error

	errManagerStopped struct{ error }
)

var (
	// ErrManagerStopped is returned by the manager implementation in this package (see also NewManager) in the case
	// that Manager.Add is attempted after the manager has already started to stop. Use errors.Is to check this case.
	ErrManagerStopped error = errManagerStopped{error: errors.New(`behaviortree.Manager.Add already stopped`)}
)

// NewManager will construct an implementation of the Manager interface, which is a stateful set of Ticker
// implementations, aggregating the behavior such that the Done channel will close when ALL tickers registered with Add
// are done, Err will return a combined error if there are any, and Stop will stop all registered tickers.
//
// Note that any error (of any registered tickers) will also trigger stopping, and stopping will prevent further
// Add calls from succeeding.
//
// As of v1.8.0, any (combined) ticker error returned by the Manager can now support error chaining (i.e. the use of
// errors.Is). Note that errors.Unwrap isn't supported, since there may be more than one. See also Manager.Err and
// Manager.Add.
func NewManager() Manager {
	result := &manager{
		done:    make(chan struct{}),
		stop:    make(chan struct{}),
		tickers: make(chan managerTicker),
	}
	return result
}

func (m *manager) Done() <-chan struct{} {
	return m.done
}

func (m *manager) Err() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.errs) != 0 {
		return errManagerTicker(m.errs)
	}
	return nil
}

func (m *manager) Stop() {
	m.once.Do(func() {
		close(m.stop)
		m.start()()
	})
}

func (m *manager) Add(ticker Ticker) error {
	if ticker == nil {
		return errors.New("behaviortree.Manager.Add nil ticker")
	}
	done := m.start()
	select {
	case <-m.stop:
	default:
		select {
		case <-m.stop:
		case m.tickers <- managerTicker{
			Ticker: ticker,
			Done:   done,
		}:
			return nil
		}
	}
	done()
	if err := m.Err(); err != nil {
		return errManagerStopped{error: err}
	}
	return ErrManagerStopped
}

func (m *manager) start() (done func()) { return m.worker.Do(m.run) }

func (m *manager) run(stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			select {
			case <-m.stop:
				select {
				case <-m.done:
				default:
					close(m.done)
				}
			default:
			}
			return
		case t := <-m.tickers:
			go m.handle(t)
		}
	}
}

func (m *manager) handle(t managerTicker) {
	select {
	case <-t.Ticker.Done():
		// note: this stop shouldn't be necessary, but has been retained for
		//       consistency, with the previous implementation)
		t.Ticker.Stop()
	case <-m.stop:
		t.Ticker.Stop()
		<-t.Ticker.Done()
	}
	if err := t.Ticker.Err(); err != nil {
		m.mu.Lock()
		m.errs = append(m.errs, err)
		m.mu.Unlock()
		m.Stop()
	}
	t.Done()
}

func (e errManagerTicker) Error() string {
	var b []byte
	for i, err := range e {
		if i != 0 {
			b = append(b, ' ', '|', ' ')
		}
		b = append(b, err.Error()...)
	}
	return string(b)
}

func (e errManagerTicker) Is(target error) bool {
	for _, err := range e {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

func (e errManagerStopped) Unwrap() error { return e.error }

func (e errManagerStopped) Is(target error) bool {
	switch target.(type) {
	case errManagerStopped:
		return true
	default:
		return false
	}
}
