# go-behaviortree

Package behaviortree provides a simple and powerful Go implementation of behavior trees without fluff.

Go doc: [https://godoc.org/github.com/joeycumines/go-behaviortree](https://godoc.org/github.com/joeycumines/go-behaviortree)

Wikipedia: [Behavior tree - AI, robotics, and control](https://en.wikipedia.org/wiki/Behavior_tree_(artificial_intelligence,_robotics_and_control))

> `go test`: 100% coverage | `go vet`: pass | `golint`: pass

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

**Features**

- Core implementation as above
- Sequence and Selector also provided as per the
    [Wikipedia page](https://en.wikipedia.org/wiki/Behavior_tree_(artificial_intelligence,_robotics_and_control))
- All, a stateless tick similar to Sequence which will attempt to run all nodes, even on (non-error) failure
- Any, a stateful tick which uses encapsulation to patch a group tick's return status to be success if any children
  succeeded, or failure if all children failed
- Async and Sync wrappers allow for the definition of time consuming logic that gets performed in serial, but without
  blocking the tick operation, and can be used to implement complex behavior such as conditional exit of running nodes
- Fork provides a group tick like Sequence and Selector, that runs it's children concurrently, and to completion
- Background, a stateful tick that supports multiple concurrent executions of _either_ static or dynamically generated
  instances of another tick implementation, where all statuses are propagated 1-1, but not necessarily in order,
  utilising the running status as it's cue to background a newly generated tick (supporting conditional backgrounding)
- RateLimit is a common building block, and implements a Tick that will return Failure as necessary to maintain a rate
- Not simply inverts a Tick's Status, but retains error handling (Failure on error value) behavior
- Implementations to actually run behavior trees are provided, and include a Ticker and Manager, but both are defined
  by interfaces and are entirely optional
- I use this implementation for several personal projects and will continue to add functionality after validating it's
  overall value (NewTickerStopOnFailure used in the example below was added in this way)

For the specific use case I built it for I have yet to find anything remotely comparable. It's also very unlikely this
will ever see a V2, but quite likely that I will be supporting V1 for years.

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
		// ticker is what actually runs this example and will tick the behavior tree defined by a single root node at
		// most once per millisecond and will stop after the first failed tick or error or context cancel
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
```
