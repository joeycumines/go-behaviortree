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
- Async and Sync wrappers allow for the definition of time consuming logic that gets performed in serial, but without
  blocking the tick operation, and can be used to implement complex behavior such as conditional exit of running nodes
- Implementations to actually run behavior trees are provided, and include a Ticker and Manager, but both are defined
  by interfaces and are entirely optional
- I use this implementation for several personal projects and will continue to add functionality after validating it's
  overall value (NewTickerStopOnFailure used in the example below was added in this way)

For the specific use case I built it for I have yet to find anything remotely comparable. It's also very unlikely this
will ever see a V2, but quite likely that I will be supporting V1 for years.

## Example Usage

The example below is straight from `example_test.go`.

TODO: more complicated example using `Async` and `Sync`

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
```
