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
	"context"
	"fmt"
	"sync"
	"time"
)

func newExampleCounter() Node {
	var (
		// counter is the shared state used by this example
		counter = 0
		// printCounter returns a node that will print the counter prefixed with the given name then succeed
		printCounter = func(name string) Node {
			return New(
				func(children []Node) (Status, error) {
					fmt.Printf("%s: %d\n", name, counter)
					return Success, nil
				},
			)
		}
		// incrementCounter is a node that will increment counter then succeed
		incrementCounter = New(
			func(children []Node) (Status, error) {
				counter++
				return Success, nil
			},
		)
	)
	return New(
		Selector, // runs each child sequentially until one succeeds (success) or all fail (failure)
		New(
			Sequence, // runs each child in order until one fails (failure) or they all succeed (success)
			New(
				func(children []Node) (Status, error) { // succeeds while counter is less than 10
					if counter < 10 {
						return Success, nil
					}
					return Failure, nil
				},
			),
			incrementCounter,
			printCounter("< 10"),
		),
		New(
			Sequence,
			New(
				func(children []Node) (Status, error) { // succeeds while counter is less than 20
					if counter < 20 {
						return Success, nil
					}
					return Failure, nil
				},
			),
			incrementCounter,
			printCounter("< 20"),
		),
	) // if both children failed (counter is >= 20) the root node will also fail
}

// ExampleNewTickerStopOnFailure_counter demonstrates the use of NewTickerStopOnFailure to implement more complex "run
// to completion" behavior using the simple modular building blocks provided by this package
func ExampleNewTickerStopOnFailure_counter() {
	// ticker is what actually runs this example and will tick the behavior tree defined by a given node at a given
	// rate and will stop after the first failed tick or error or context cancel
	ticker := NewTickerStopOnFailure(
		context.Background(),
		time.Millisecond,
		newExampleCounter(),
	)
	// waits until ticker stops, which will be on the first failure of it's root node
	<-ticker.Done()
	// every Tick may return an error which would automatically cause a failure and propagation of the error
	if err := ticker.Err(); err != nil {
		panic(err)
	}
	// Output:
	// < 10: 1
	// < 10: 2
	// < 10: 3
	// < 10: 4
	// < 10: 5
	// < 10: 6
	// < 10: 7
	// < 10: 8
	// < 10: 9
	// < 10: 10
	// < 20: 11
	// < 20: 12
	// < 20: 13
	// < 20: 14
	// < 20: 15
	// < 20: 16
	// < 20: 17
	// < 20: 18
	// < 20: 19
	// < 20: 20
}

// ExampleMemorize_cancellationWithContextCancel demonstrates how support for reactive logic that uses context can
// be implemented
func ExampleMemorize_cancellationWithContextCancel() {
	type Str string
	var (
		ctx    context.Context
		cancel context.CancelFunc
		debug  = func(label string, tick Tick) Tick {
			return func(children []Node) (status Status, err error) {
				status, err = tick(children)
				fmt.Printf("%s returned (%v, %v)\n", label, status, err)
				return
			}
		}
		recorded = func(statuses ...Status) Tick {
			return func([]Node) (status Status, err error) {
				status = statuses[0]
				statuses = statuses[1:]
				return
			}
		}
		counter int
		ticker  = NewTickerStopOnFailure(
			context.Background(),
			time.Millisecond,
			New(
				All,
				New(
					Memorize(debug(`memorized`, All)),
					New(func([]Node) (Status, error) {
						counter++
						ctx, cancel = context.WithCancel(context.WithValue(context.Background(), Str(`n`), counter))
						return Success, nil
					}), // prepare the context
					New(
						debug(`sequence`, Sequence),
						New(debug(`guard`, recorded(
							Success,
							Success,
							Success,
							Success,
							Failure,
						))),
						New(func([]Node) (Status, error) {
							fmt.Printf("[start action] context #%d's err=%v\n", ctx.Value(Str(`n`)), ctx.Err())
							return Success, nil
						}),
						New(debug(`action`, recorded(
							Running,
							Running,
							Success,
							Running,
						))),
					),
					New(func([]Node) (Status, error) {
						cancel()
						return Success, nil
					}), // cancel the context
				),
				New(func([]Node) (Status, error) {
					fmt.Printf("[end memorized] context #%d's err=%v\n", ctx.Value(Str(`n`)), ctx.Err())
					return Success, nil
				}),
			),
		)
	)

	<-ticker.Done()
	if err := ticker.Err(); err != nil {
		panic(err)
	}
	//output:
	//guard returned (success, <nil>)
	//[start action] context #1's err=<nil>
	//action returned (running, <nil>)
	//sequence returned (running, <nil>)
	//memorized returned (running, <nil>)
	//guard returned (success, <nil>)
	//[start action] context #1's err=<nil>
	//action returned (running, <nil>)
	//sequence returned (running, <nil>)
	//memorized returned (running, <nil>)
	//guard returned (success, <nil>)
	//[start action] context #1's err=<nil>
	//action returned (success, <nil>)
	//sequence returned (success, <nil>)
	//memorized returned (success, <nil>)
	//[end memorized] context #1's err=context canceled
	//guard returned (success, <nil>)
	//[start action] context #2's err=<nil>
	//action returned (running, <nil>)
	//sequence returned (running, <nil>)
	//memorized returned (running, <nil>)
	//guard returned (failure, <nil>)
	//sequence returned (failure, <nil>)
	//memorized returned (failure, <nil>)
	//[end memorized] context #2's err=context canceled
}

// ExampleBackground_asyncJobQueue implements a basic example of backgrounding of long-running tasks that may be
// performed concurrently, see ExampleNewTickerStopOnFailure_counter for an explanation of the ticker
func ExampleBackground_asyncJobQueue() {
	type (
		Job struct {
			Name     string
			Duration time.Duration
			Done     chan struct{}
		}
	)
	var (
		// doWorker performs the actual "work" for a Job
		doWorker = func(job Job) {
			fmt.Printf("[worker] job \"%s\" STARTED\n", job.Name)
			time.Sleep(job.Duration)
			fmt.Printf("[worker] job \"%s\" FINISHED\n", job.Name)
			close(job.Done)
		}
		// queue be sent jobs, which will be received within the ticker
		queue = make(chan Job, 50)
		// doClient sends and waits for a job
		doClient = func(name string, duration time.Duration) {
			job := Job{name, duration, make(chan struct{})}
			ts := time.Now()
			fmt.Printf("[client] job \"%s\" STARTED\n", job.Name)
			queue <- job
			<-job.Done
			fmt.Printf("[client] job \"%s\" FINISHED\n", job.Name)
			t := time.Since(ts)
			d := t - job.Duration
			if d < 0 {
				d *= -1
			}
			if d > time.Millisecond*50 {
				panic(fmt.Errorf(`job "%s" expected %s actual %s`, job.Name, job.Duration.String(), t.String()))
			}
		}
		// running keeps track of the number of running jobs
		running = func() func(delta int64) int64 {
			var (
				value int64
				mutex sync.Mutex
			)
			return func(delta int64) int64 {
				mutex.Lock()
				defer mutex.Unlock()
				value += delta
				return value
			}
		}()
		// done will be closed when it's time to exit the ticker
		done   = make(chan struct{})
		ticker = NewTickerStopOnFailure(
			context.Background(),
			time.Millisecond,
			New(
				Sequence,
				New(func(children []Node) (Status, error) {
					select {
					case <-done:
						return Failure, nil
					default:
						return Success, nil
					}
				}),
				func() Node {
					// the tick is initialised once, and is stateful (though the tick it's wrapping isn't)
					tick := Background(func() Tick { return Selector })
					return func() (Tick, []Node) {
						// this block will be refreshed each time that a new job is started
						var (
							job Job
						)
						return tick, []Node{
							New(
								Sequence,
								Sync([]Node{
									New(func(children []Node) (Status, error) {
										select {
										case job = <-queue:
											running(1)
											return Success, nil
										default:
											return Failure, nil
										}
									}),
									New(Async(func(children []Node) (Status, error) {
										defer running(-1)
										doWorker(job)
										return Success, nil
									})),
								})...,
							),
							// no job available - success
							New(func(children []Node) (Status, error) {
								return Success, nil
							}),
						}
					}
				}(),
			),
		)
		wg sync.WaitGroup
	)
	wg.Add(1)
	run := func(name string, duration time.Duration) {
		wg.Add(1)
		defer wg.Done()
		doClient(name, duration)
	}

	fmt.Printf("running jobs: %d\n", running(0))

	go run(`1. 120ms`, time.Millisecond*120)
	time.Sleep(time.Millisecond * 25)
	go run(`2. 70ms`, time.Millisecond*70)
	time.Sleep(time.Millisecond * 25)
	fmt.Printf("running jobs: %d\n", running(0))

	doClient(`3. 150ms`, time.Millisecond*150)
	time.Sleep(time.Millisecond * 50)
	fmt.Printf("running jobs: %d\n", running(0))

	time.Sleep(time.Millisecond * 50)
	wg.Done()
	wg.Wait()
	close(done)
	<-ticker.Done()
	if err := ticker.Err(); err != nil {
		panic(err)
	}
	//output:
	//running jobs: 0
	//[client] job "1. 120ms" STARTED
	//[worker] job "1. 120ms" STARTED
	//[client] job "2. 70ms" STARTED
	//[worker] job "2. 70ms" STARTED
	//running jobs: 2
	//[client] job "3. 150ms" STARTED
	//[worker] job "3. 150ms" STARTED
	//[worker] job "2. 70ms" FINISHED
	//[client] job "2. 70ms" FINISHED
	//[worker] job "1. 120ms" FINISHED
	//[client] job "1. 120ms" FINISHED
	//[worker] job "3. 150ms" FINISHED
	//[client] job "3. 150ms" FINISHED
	//running jobs: 0
}

// ExampleContext demonstrates how the Context implementation may be used to integrate with the context package
func ExampleContext() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var (
		btCtx = new(Context).WithTimeout(ctx, time.Millisecond*100)
		debug = func(args ...interface{}) Tick {
			return func([]Node) (Status, error) {
				fmt.Println(args...)
				return Success, nil
			}
		}
		counter      int
		counterEqual = func(v int) Tick {
			return func([]Node) (Status, error) {
				if counter == v {
					return Success, nil
				}
				return Failure, nil
			}
		}
		counterInc Tick = func([]Node) (Status, error) {
			counter++
			//fmt.Printf("counter = %d\n", counter)
			return Success, nil
		}
		ticker = NewTicker(ctx, time.Millisecond, New(
			Sequence,
			New(
				Selector,
				New(Not(btCtx.Err)),
				New(
					Sequence,
					New(debug(`(re)initialising btCtx...`)),
					New(btCtx.Init),
					New(Not(btCtx.Err)),
				),
			),
			New(
				Selector,
				New(
					Sequence,
					New(counterEqual(0)),
					New(debug(`blocking on context-enabled tick...`)),
					New(
						btCtx.Tick(func(ctx context.Context, children []Node) (Status, error) {
							fmt.Printf("NOTE children (%d) passed through\n", len(children))
							<-ctx.Done()
							return Success, nil
						}),
						New(Sequence),
						New(Sequence),
					),
					New(counterInc),
				),
				New(
					Sequence,
					New(counterEqual(1)),
					New(debug(`blocking on done...`)),
					New(btCtx.Done),
					New(counterInc),
				),
				New(
					Sequence,
					New(counterEqual(2)),
					New(debug(`canceling local then rechecking the above...`)),
					New(btCtx.Cancel),
					New(btCtx.Err),
					New(btCtx.Tick(func(ctx context.Context, children []Node) (Status, error) {
						<-ctx.Done()
						return Success, nil
					})),
					New(btCtx.Done),
					New(counterInc),
				),
				New(
					Sequence,
					New(counterEqual(3)),
					New(debug(`canceling parent then rechecking the above...`)),
					New(func([]Node) (Status, error) {
						cancel()
						return Success, nil
					}),
					New(btCtx.Err),
					New(btCtx.Tick(func(ctx context.Context, children []Node) (Status, error) {
						<-ctx.Done()
						return Success, nil
					})),
					New(btCtx.Done),
					New(debug(`exiting...`)),
				),
			),
		))
	)

	<-ticker.Done()

	//output:
	//(re)initialising btCtx...
	//blocking on context-enabled tick...
	//NOTE children (2) passed through
	//(re)initialising btCtx...
	//blocking on done...
	//(re)initialising btCtx...
	//canceling local then rechecking the above...
	//(re)initialising btCtx...
	//canceling parent then rechecking the above...
	//exiting...
}
