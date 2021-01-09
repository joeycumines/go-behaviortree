[![Go Report Card](https://goreportcard.com/badge/joeycumines/go-behaviortree)](https://goreportcard.com/report/joeycumines/go-behaviortree)

# go-behaviortree

Package behaviortree provides a simple and powerful Go implementation of behavior trees without fluff.

Go doc: [https://godoc.org/github.com/joeycumines/go-behaviortree](https://godoc.org/github.com/joeycumines/go-behaviortree)

Wikipedia: [Behavior tree - AI, robotics, and control](https://en.wikipedia.org/wiki/Behavior_tree_(artificial_intelligence,_robotics_and_control))

```go
type (
	// Node represents an node in a tree, that can be ticked
	Node func() (Tick, []Node)

	// Tick represents the logic for a node, which may or may not be stateful
	Tick func(children []Node) (Status, error)

	// Status is a type with three valid values, Running, Success, and Failure, the three possible states for BTs
	Status int
)

// Tick runs the node's tick function with it's children
func (n Node) Tick() (Status, error)
```

## Features

- Core behavior tree implementation (the types above + `Sequence` and `Selector`)
- Tools to aide implementation of "reactive" behavior trees (`Memorize`, `Async`, `Sync`)
- Implementations to run and manage behavior trees (`NewManager`, `NewTicker`)
- Collection of `Tick` implementations / wrappers (targeting various use cases)
- Context-like mechanism to attach metadata to `Node` values that can transit API boundaries / encapsulation
- Basic tree debugging capabilities via implementation of `fmt.Stringer` (see also `DefaultPrinter`, `Node.Frame`)
- Experimental support for the PA-BT planning algorithm via [github.com/joeycumines/go-pabt](https://github.com/joeycumines/go-pabt)

## Design

This library provides an implementation of the generalised behavior tree pattern. The three types above along with the
stateless `Tick` functions `Selector` and `Sequence` (line for line from Wikipedia) make up the entirety of the core
functionality. Additional features have been derived by means of iterating against specific, real-world problem cases.

It is recommended that anyone interested in understanding behavior trees read from the start of chapter 1 of
[Colledanchise, Michele & Ogren, Petter. (2018). Behavior Trees in Robotics and AI: An Introduction. 10.1201/9780429489105.](https://www.researchgate.net/publication/319463746_Behavior_Trees_in_Robotics_and_AI_An_Introduction)
until at least (the end of) `1.3.2 - Control Flow Nodes with Memory`. Further context should (hopefully) go a long way
in aiding newcomers to gain a firm grasp of the relevant behavior and associated patterns, in order to start building
automations using behavior trees.

N.B. Though it does appear to be an excellent resource, _Behavior Trees in Robotics and AI: An Introduction._ was only
     brought to my attention well after the release of `v1.3.0`. In particular, concepts and terms are unlikely to be
     entirely consistent.

### Reactivity

Executing a behavior tree sequentially (i.e. without the use of nodes that return `Running`) can be an effective way
to decompose complicated switching logic. Though effective, this design suffers from limitations common to traditional
finite state machines, which tends to cripple the modularity of any interruptable operations, as each tick must manage
it's own exit condition(s). Golang's `context` doesn't really help in that case, either, being a communication
mechanism, rather than a control mechanism. As an alternative, implementations may use pre-conditions (preceding
child(ren) in a sequence), guarding an asynchronous tick. Context support may be desirable, and may be implemented as
`Tick` implementations(s). This package provides a `Context` implementation to address several common use cases, such
as operations with timeouts. Context support is peripheral in that it's only relevant to a subset of implementations,
and was deliberately omitted from the core `Tick` type.

### Modularity

This library **only** concerns itself with the task of actually running behavior trees. It is deliberately designed
to make it straightforward to plug in external implementations. External implementations _may_ be integrated as a
fully-fledged recursively generated tree of `Node`. Tree-aware debug tracing could be implemented using a similar
mechanism. At the end of the day though, implementations just need to `Tick`.

## Implementation

This section attempts to provide better insight into the _intent_ of certain implementation details. Part of this
involves offering implementation **guidelines**, based on _imagined_ use cases (beyond those explored as part of each
implementation). It ain't gospel.

### Executing BTs

#### Use explicit exit conditions

- Return `error` to handle failures that should terminate tree execution
- BTs that "run until completion" may be implemented using `NewTickerStopOnFailure` (internally it just returns
  an error then strips any occurrence of that that error from the ticker's result)
- Panics may be used as normal, and are suitable for cases such unrecoverable errors that _shouldn't_ happen

### Shared state

#### `Tick` implementations should be grouped in a way that makes sense, in the context of any shared state

- It may be convenient to expose a group of `Tick` prototypes with shared state as methods of a struct
- Global state should be avoided for all the regular reasons, but especially since it defeats a lot of the point
  of having modular, composable behavior

#### Encapsulate `Tick` implementations, rather than `Node`

- Children may be may be modified, but only until the (outer) `Tick` returns it's next non-running status
- This mechanism _theoretically_ facilitates dynamically generated trees while simultaneously supporting more complex
  / concurrency-heavy implementations made up of reusable building blocks

## Roadmap

I am actively maintaining this project, and will be for the foreseeable future. It has been "feature complete" for
some time though, so additional functionality will assessed on a case-by-case basis.

## Example Usage

The examples below are straight from `example_test.go`.

```go
// ExampleNewTickerStopOnFailure_counter demonstrates the use of NewTickerStopOnFailure to implement more complex "run
// to completion" behavior using the simple modular building blocks provided by this package
func ExampleNewTickerStopOnFailure_counter() {
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
		// ticker is what actually runs this example and will tick the behavior tree defined by a given node at a given
		// rate and will stop after the first failed tick or error or context cancel
		ticker = NewTickerStopOnFailure(
			context.Background(),
			time.Millisecond,
			New(
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
			), // if both children failed (counter is >= 20) the root node will also fail
		)
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
						ctx, cancel = context.WithCancel(context.WithValue(context.Background(), `n`, counter))
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
							fmt.Printf("[start action] context #%d's err=%v\n", ctx.Value(`n`), ctx.Err())
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
					fmt.Printf("[end memorized] context #%d's err=%v\n", ctx.Value(`n`), ctx.Err())
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
			t := time.Now().Sub(ts)
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
```
