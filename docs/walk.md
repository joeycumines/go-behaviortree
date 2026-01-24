# Walk - Tree Introspection

The `Walk` function provides a mechanism for depth-first traversal of a behavior tree, serving as a fundamental building block for introspection tools such as debuggers, visualizers, and static analyzers.

## Overview

Unlike a running tree which executes `Tick` logic, `Walk` is designed to inspect the *structure* of the tree. It traverses the hierarchy of nodes, respecting logical structure metadata transparently where provided.

### Signature

```go
package behaviortree

func Walk(n Metadata, fn func(n Metadata) bool)
```

- **n**: The root object to start traversal from. This is typically a `Node` (which implements `Metadata`), but can be any implementation of the `Metadata` interface.
- **fn**: A callback function executed for every visited node. Returning `false` stops traversal.

## Mechanics

Breadth-first or depth-first? **Depth-first**.
The walker visits the current node `n`, then recursively visits each valid child yielded by `n.Children()`.

### Structural Resolution

The `Walk` function is designed to traverse the "conceptual" tree structure, which may differ from the "physical" structure (the actual compiled closures and function pointers).

This decoupling is achieved via the `Metadata` interface:

```go
package behaviortree

type Metadata interface {
	Value(key any) any
	Children(yield func(Metadata) bool)
}

```

When `Walk` visits a node `n`, it effectively iterates over `n.Children()`. The `Node` implementation of this method resolves the children in the following order:

1. **Conceptual Structure (Metadata)**:
   It first checks `n.Structure()` (accessed via `Value`).
    * If this returns a non-nil sequence of `Metadata` items, `Walk` iterates over these items *instead* of physically expanding the node.
    * This allows for "virtualized" subtrees. For example, a complex `Selector` could present itself to the walker as a simple leaf, or a leaf could generate a sequence of virtual nodes representing its internal state.
    * **Efficiency Note**: By yielding objects that strictly implement `Metadata` (and aren't necessarily full `Node` instances), one can avoid the overhead of the `Node` machinery (specifically the `Value` locking mechanism) for large, read-only subtrees.

2. **Physical Structure (Expansion)**:
   If `Structure()` returns `nil` (the default), the node falls back to expanding itself.
    * It executes `tick, children := n()`.
    * This uses the standard `behaviortree` factory pattern to retrieve the actual child nodes.
    * This ensures that by default, `Walk` accurately reflects the execution hierarchy.

## Performance Considerations

### Cost of Traversal

The cost of `Walk` is linear with respect to the number of nodes in the tree ($O(N)$), provided `Structure()` and node expansion are constant time operations.

* **Node Expansion**: Since `Walk` must execute `n()` to discover children for standard nodes, the performance depends on the cost of these factory functions. In idiomatic `behaviortree` usage, these are lightweight closures returning pre-allocated slices.
* **Metadata Access**: Accessing `Structure()` involves the `Node.Value` mechanism, which uses `sync.Mutex` and atomic checks. This adds a constant overhead per node.

### Benchmarks

The following benchmarks verify the performance characteristics of `Walk` on an Apple M2 Pro, comparing standard nodes vs nodes utilizing `Structure` metadata.

| Benchmark               | Iterations | Time/Op   | Alloc/Op | Notes                                    |
|:------------------------|:-----------|:----------|:---------|:-----------------------------------------|
| `Walk_Deep100`          | 48,880     | ~22.6 µs  | 12.9 KB  | Linear depth traversal                   |
| `Walk_Wide100`          | 54,868     | ~21.7 µs  | 12.9 KB  | Breadth traversal (flat)                 |
| `Walk_LargeTree`        | 6,991      | ~171.3 µs | 99.9 KB  | Mixed (781 nodes)                        |
| `Walk_StructureDeep100` | 4,012      | ~300.3 µs | 217.7 KB | **~13x Slower** due to metadata overhead |

**Analysis**:

* **Standard Traversal**: Highly efficient (~200ns per node). Memory allocation is dominated by the recursive stack and slice handling.
* **Structure Overhead**: Using `WithStructure` introduces significant overhead (approx. 13x slower in deep trees). This is due to the `Node.Value` synchronization mechanism (`sync.Mutex` and channel ops) required for each node visit.
* **Recommendation**: Use `WithStructure` judiciously for complex composite nodes where logical abstraction is valuable, rather than universally on all leaf nodes.

### Concurrency Safety

`Walk` is **not** safe to call concurrently on a tree that is being mutated, although `behaviortree` nodes are typically immutable after construction.

> **Note on Deadlocks**: The global lock used by `Node.Value` is not re-entrant. While standard `Walk` usage is safe, custom `Node` implementations that recursively call `Value` during their own definition phase (inside `n()`) will deadlock.
>
> **Performance Tip**: To avoid the contention of the global `Value` lock entirely for a subtree, implement a custom `Metadata` type that is *not* a `Node` and return it via `Structure()`. This allows `Walk` to traverse that subtree without touching the `behaviortree` lock mechanism.

## Best Practices

* **Purity**: Ensure `Node` functions used with `Walk` do not contain side effects. `Walk` executes them.
* **Avoid Cycles**: `Walk` does not detect cycles. Ensure your tree is a Directed Acyclic Graph (DAG) to prevent infinite recursion.
* **Use `WithStructure`**: Use `WithStructure` to provide meaningful debug hierarchies for complex composite nodes, keeping technical implementation details hidden from high-level visualizers.
