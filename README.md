# go-behaviortree

Package behaviortree provides a functional style implementation of behavior trees
in Golang with no fluff.

Go doc: [https://godoc.org/github.com/joeycumines/go-behaviortree](https://godoc.org/github.com/joeycumines/go-behaviortree)

Wikipedia: [Behavior tree - AI, robotics, and control](https://en.wikipedia.org/wiki/Behavior_tree_(artificial_intelligence,_robotics_and_control))

```go
type (
	// Node represents an node in a tree, that can be ticked
	Node func() (Tick, []Node)

	// Tick represents the logic for a node, which may or may not be stateful
	Tick func(children []Node) (Status, error)
)

// Tick runs the node's tick function with it's children
func (n Node) Tick() (Status, error)
```

- Core implementation as above
- Sequence and Selector also provided as per the
    [Wikipedia page](https://en.wikipedia.org/wiki/Behavior_tree_(artificial_intelligence,_robotics_and_control))
- Async and Sync wrappers allow for the definition of time consuming logic that gets performed 
    in serial, but without blocking the tick operation.

Examples to come at some point.
