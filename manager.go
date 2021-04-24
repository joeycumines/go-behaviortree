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
	"strings"
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
		mutex   sync.Mutex
		tickers []Ticker
		errs    []string
		stopped bool
		done    chan struct{}
	}
)

// NewManager will construct an implementation of the Manager interface, which is a stateful set of Ticker
// implementations, aggregating the behavior such that the Done channel will close when ALL tickers registered with Add
// are done, Err will return a combined error if there are any, and Stop will stop all registered tickers.
//
// Note that any error (of any registered tickers) will also trigger stopping, and stopping will prevent further
// Add calls from succeeding.
func NewManager() Manager {
	result := &manager{
		done: make(chan struct{}),
	}
	return result
}

func (m *manager) Done() <-chan struct{} {
	return m.done
}

func (m *manager) Err() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.err()
}

func (m *manager) Stop() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.stopOnce()
}

func (m *manager) Add(ticker Ticker) error {
	if ticker == nil {
		return errors.New("behaviortree.Manager.Add nil ticker")
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.check()
	if m.stopped {
		if err := m.err(); err != nil {
			return err
		}
		return errors.New("behaviortree.Manager.Add already stopped")
	}
	m.tickers = append(m.tickers, ticker)
	go func() {
		<-ticker.Done()
		m.mutex.Lock()
		defer m.mutex.Unlock()
		m.check()
	}()
	return nil
}

func (m *manager) err() error {
	if len(m.errs) != 0 {
		return errors.New(strings.Join(m.errs, " | "))
	}
	return nil
}

func (m *manager) stopOnce() {
	if !m.stopped {
		m.stopped = true
		go m.cleanup()
	}
}

func (m *manager) finish(i int) {
	m.tickers[i].Stop()
	<-m.tickers[i].Done()
	if err := m.tickers[i].Err(); err != nil {
		m.errs = append(m.errs, err.Error())
		m.stopOnce()
	}
	m.tickers[i] = m.tickers[len(m.tickers)-1]
	m.tickers[len(m.tickers)-1] = nil
	m.tickers = m.tickers[:len(m.tickers)-1]
}

func (m *manager) check() {
	for i := 0; i < len(m.tickers); i++ {
		select {
		case <-m.tickers[i].Done():
			m.finish(i)
			i--
		default:
		}
	}
}

func (m *manager) cleanup() {
	m.mutex.Lock()
	for i := 0; i < len(m.tickers); i++ {
		m.finish(i)
		i--
	}
	close(m.done)
	m.mutex.Unlock()
}
