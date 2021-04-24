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
	"context"
	"errors"
	"sync"
	"time"
)

type (
	// Ticker models a node runner
	Ticker interface {
		// Done will close when the ticker is fully stopped.
		Done() <-chan struct{}

		// Err will return any error that occurs.
		Err() error

		// Stop shutdown the ticker asynchronously.
		Stop()
	}

	// tickerCore is the base ticker implementation
	tickerCore struct {
		ctx    context.Context
		cancel context.CancelFunc
		node   Node
		ticker *time.Ticker
		done   chan struct{}
		stop   chan struct{}
		once   sync.Once
		mutex  sync.Mutex
		err    error
	}

	// tickerStopOnFailure is an implementation of a ticker that will run until the first error
	tickerStopOnFailure struct {
		Ticker
	}
)

var (
	// errExitOnFailure is a specific error used internally to exit tickers constructed with NewTickerStopOnFailure,
	// and won't be returned by the tickerStopOnFailure implementation
	errExitOnFailure = errors.New("errExitOnFailure")
)

// NewTicker constructs a new Ticker, which simply uses time.Ticker to tick the provided node periodically, note
// that a panic will occur if ctx is nil, duration is <= 0, or node is nil.
//
// The node will tick until the first error or Ticker.Stop is called, or context is canceled, after which any error
// will be made available via Ticker.Err, before closure of the done channel, indicating that all resources have been
// freed, and any error is available.
func NewTicker(ctx context.Context, duration time.Duration, node Node) Ticker {
	if ctx == nil {
		panic(errors.New("behaviortree.NewTicker nil context"))
	}

	if duration <= 0 {
		panic(errors.New("behaviortree.NewTicker duration <= 0"))
	}

	if node == nil {
		panic(errors.New("behaviortree.NewTicker nil node"))
	}

	result := &tickerCore{
		node:   node,
		ticker: time.NewTicker(duration),
		done:   make(chan struct{}),
		stop:   make(chan struct{}),
	}

	result.ctx, result.cancel = context.WithCancel(ctx)

	go result.run()

	return result
}

// NewTickerStopOnFailure returns a new Ticker that will exit on the first Failure, but won't return a non-nil Err
// UNLESS there was an actual error returned, it's built on top of the same core implementation provided by NewTicker,
// and uses that function directly, note that it will panic if the node is nil, the panic cases for NewTicker also
// apply.
func NewTickerStopOnFailure(ctx context.Context, duration time.Duration, node Node) Ticker {
	if node == nil {
		panic(errors.New("behaviortree.NewTickerStopOnFailure nil node"))
	}

	return tickerStopOnFailure{
		Ticker: NewTicker(
			ctx,
			duration,
			func() (Tick, []Node) {
				tick, children := node()
				if tick == nil {
					return nil, children
				}
				return func(children []Node) (Status, error) {
					status, err := tick(children)
					if err == nil && status == Failure {
						err = errExitOnFailure
					}
					return status, err
				}, children
			},
		),
	}
}

func (t *tickerCore) run() {
	var err error
TickLoop:
	for err == nil {
		select {
		case <-t.ctx.Done():
			err = t.ctx.Err()
			break TickLoop
		case <-t.stop:
			break TickLoop
		case <-t.ticker.C:
			_, err = t.node.Tick()
		}
	}
	t.mutex.Lock()
	t.err = err
	t.mutex.Unlock()
	t.Stop()
	t.cancel()
	close(t.done)
}

func (t *tickerCore) Done() <-chan struct{} {
	return t.done
}

func (t *tickerCore) Err() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.err
}

func (t *tickerCore) Stop() {
	t.once.Do(func() {
		t.ticker.Stop()
		close(t.stop)
	})
}

func (t tickerStopOnFailure) Err() error {
	err := t.Ticker.Err()
	if err == errExitOnFailure {
		return nil
	}
	return err
}
