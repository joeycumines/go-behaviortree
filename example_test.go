package behaviortree

import (
	"context"
	"fmt"
	"time"
)

func ExampleSimpleTicker() {
	var (
		counter                    = 0
		nodeGuardCounterLessThan10 = NewNode(
			func(children []Node) (Status, error) {
				if counter < 10 {
					return Success, nil
				}
				return Failure, nil
			},
			nil,
		)
		nodeGuardCounterLessThan20 = NewNode(
			func(children []Node) (Status, error) {
				if counter < 20 {
					return Success, nil
				}
				return Failure, nil
			},
			nil,
		)
		newNodePrintCounter = func(name string) Node {
			return NewNode(
				func(children []Node) (Status, error) {
					fmt.Printf("%s: %d\n", name, counter)
					return Success, nil
				},
				nil,
			)
		}
		nodeIncrementCounter = NewNode(
			func(children []Node) (Status, error) {
				counter++
				return Success, nil
			},
			nil,
		)
		nodeRoot = NewNode(
			Selector,
			[]Node{
				NewNode(
					Sequence,
					[]Node{
						nodeGuardCounterLessThan10,
						nodeIncrementCounter,
						newNodePrintCounter("< 10"),
					},
				),
				NewNode(
					Sequence,
					[]Node{
						nodeGuardCounterLessThan20,
						nodeIncrementCounter,
						newNodePrintCounter("< 20"),
					},
				),
			},
		)
		tickerRoot = NewTickerStopOnFailure(context.Background(), time.Millisecond, nodeRoot)
	)
	<-tickerRoot.Done()
	if err := tickerRoot.Err(); err != nil {
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
