# Walk - Tree Introspection

The `Walk` function provides a mechanism for depth-first traversal of a behavior tree, serving as a fundamental building block for introspection tools such as debuggers, visualizers, and static analyzers.

## Overview

Unlike a running tree which executes `Tick` logic, `Walk` is designed to inspect the *structure* of the tree. It traverses the hierarchy of nodes, respecting logical structure metadata transparently where provided.

### Signature

```go
func Walk(n Node, fn func (n Node))
```

- **n**: The root node to start traversal from.
- **fn**: A callback function executed for every visited node (including the root).

## Mechanics

Breadth-first or depth-first? **Depth-first**.
The walker visits the current node `n`, then recursively visits each valid child.

### Structural Resolution

A critical feature of `Walk` is its ability to distinguish between "physical" structure (the actual functions and closures implementing the tree) and "logical" structure (what the user conceptually considers the tree).

For any given node `n`, `Walk` determines children in the following order:

1. **Logical Structure (Metadata)**:
   It first checks `n.Structure()`. If this returns a non-nil slice, `Walk` iterates over these nodes.
    * This allows decorators, wrappers, or complex leaf nodes (like FSMs) to present a simplified or virtual hierarchy to tools.
    * An empty slice returned by `Structure()` effectively "masks" the node's children, making it appear as a leaf to the walker.

2. **Physical Structure (Expansion)**:
   If `Structure()` returns `nil` (default), `Walk` falls back to expanding the node using its definition.
    * It executes `tick, children := n()`.
    * This executes the node's factory function to retrieve its children.
    * CRITICAL: This assumes the node definition `n()` is a side-effect-free factory, which is the standard pattern for `behaviortree`.

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

The `Node.Value` mechanism used by `Structure()` relies on a global lock (`valueCallMutex`) to ensure safety during the introspection step. This prevents data races but means concurrent `Walk` calls (or `Walk` concurrent with other introspection) will be serialized globally.

> **Note on Deadlocks**: The global lock is not re-entrant. While standard `Walk` usage is strictly sequential and safe, custom `Node` implementations that perform recursive introspection (calling `Value` inside their own definition) will deadlock. This is an intentional constraint to enforce simple, side-effect-free structural definitions.

## Best Practices

* **Purity**: Ensure `Node` functions used with `Walk` do not contain side effects. `Walk` executes them.
* **Avoid Cycles**: `Walk` does not detect cycles. Ensure your tree is a Directed Acyclic Graph (DAG) to prevent infinite recursion.
* **Use `WithStructure`**: Use `WithStructure` to provide meaningful debug hierarchies for complex composite nodes, keeping technical implementation details hidden from high-level visualizers.
