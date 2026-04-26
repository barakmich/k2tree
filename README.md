# K2Tree

Compressed graph adjacency matrix using the [K²-tree](https://doi.org/10.1007/978-3-642-03784-9_3) data structure.

Two implementations: **Go** (original) and **Rust** (port). Both share the same core algorithm with language-appropriate optimizations.

## What is a K²-tree?

A K²-tree is a hierarchical, compressed representation of a binary matrix (graph adjacency matrix). It recursively partitions the matrix into quadrants, storing presence/absence of edges as bits. This yields extremely space-efficient storage for sparse graphs while supporting efficient neighbor iteration.

**Key properties:**

- **Space-efficient:** Stores sparse graphs using ~1-4 bits per edge (vs 64+ bits per edge in adjacency lists)
- **Read-optimized:** Once built, iteration over neighbors is fast and cache-friendly
- **Static after build:** Optimized for workloads where the graph is built once, then queried many times

**Use cases:** Social network analysis, web graph storage, any domain with large sparse graphs.

## Quick Start

### Rust

```toml
# Cargo.toml
[dependencies]
k2tree = { path = "rust" }
```

```rust
use k2tree::{K2Tree, SliceArray, BitArray};

let mut tree = K2Tree::new(SliceArray::new(), SliceArray::new());

// Add edges (zero-indexed nodes)
tree.add(0, 1).unwrap();
tree.add(0, 2).unwrap();
tree.add(1, 2).unwrap();

// Iterate outgoing edges from node 0
let neighbors = tree.from(0).extract_all();
// neighbors = [1, 2]

// Iterate incoming edges to node 2
let predecessors = tree.to(2).extract_all();
// predecessors = [0, 1]
```

### Go

```go
import "github.com/barakmich/k2tree"

k2, _ := k2tree.New()
k2.Add(0, 1)
k2.Add(0, 2)
k2.Add(1, 2)

// Outgoing edges from node 0
for it := k2.From(0); it.Next(); {
    fmt.Println(it.Value()) // 1, 2
}

// Incoming edges to node 2
for it := k2.To(2); it.Next(); {
    fmt.Println(it.Value()) // 0, 1
}
```

## API

### K2Tree

| Method                                                            | Description                              |
| ----------------------------------------------------------------- | ---------------------------------------- |
| `New()` / `K2Tree::new(tbits, lbits)`                             | Create with default config (16×16)       |
| `NewWithConfig(config)` / `new_with_config(tbits, lbits, config)` | Create with custom config                |
| `Add(i, j)` / `add(i, j)`                                         | Assert edge from node i to node j        |
| `From(i)` / `from(i)`                                             | Iterator over outgoing edges from node i |
| `To(j)` / `to(j)`                                                 | Iterator over incoming edges to node j   |
| `Stats()` / `stats()`                                             | Memory usage statistics                  |

### Iterator

| Method                           | Description                                       |
| -------------------------------- | ------------------------------------------------- |
| `Next()` / `next_edge()`         | Advance to next neighbor, returns `true` if found |
| `Value()` / `value()`            | Current neighbor node index                       |
| `ExtractAll()` / `extract_all()` | Collect all remaining neighbors into a slice/vec  |

### BitArray backends

The K2Tree uses two `BitArray` instances internally (`tbits` for tree levels, `lbits` for the leaf level). Different backends trade off memory, speed, and features:

| Backend           | Description                        | Go  | Rust |
| ----------------- | ---------------------------------- | --- | ---- |
| `SliceArray`      | Simple vector-backed bit array     | ✅  | ✅   |
| `LruArray`        | LRU cache over popcount operations | ✅  | ✅   |
| `QuartileIndex`   | 4-way partitioned popcount cache   | ✅  | ✅   |
| `PagedBitarray`   | Multi-level paged bit array        | ✅  | ❌   |
| `PagedSliceArray` | Multiple slice arrays in pages     | ✅  | ❌   |
| `Pagefile`        | Memory-mapped file persistence     | ✅  | ❌   |

### Configuration

K²-tree behavior is controlled by `Config`, which defines the tree and leaf layer sizes:

| Config                   | Tree layers   | Leaf layer    | Best for                            |
| ------------------------ | ------------- | ------------- | ----------------------------------- |
| `FOUR_FOUR_CONFIG`       | 4 bits (2×2)  | 4 bits (2×2)  | Very sparse graphs, max compression |
| `SIXTEEN_FOUR_CONFIG`    | 16 bits (4×4) | 4 bits (2×2)  | Default, good balance               |
| `SIXTEEN_SIXTEEN_CONFIG` | 16 bits (4×4) | 16 bits (4×4) | Denser graphs, faster iteration     |

The `kPerLayer` value (2, 4, 8, 16) controls how the matrix is partitioned at each level. Larger values mean fewer tree levels but wider nodes.

## Performance

### Benchmarks

Run Go benchmarks:

```bash
go test -bench=. -benchmem
```

Run Rust benchmarks:

```bash
cd rust
cargo bench --bench k2tree_bench
```

Run iteration benchmarks (larger datasets, smaller sample size):

```bash
cd rust
cargo bench --bench k2tree_bench -- iter_benches
```

### Guidelines

- **For most workloads:** `SIXTEEN_FOUR_CONFIG` with `SliceArray` is the sweet spot
- **For very sparse graphs:** `FOUR_FOUR_CONFIG` gives better compression
- **For dense graphs:** `SIXTEEN_SIXTEEN_CONFIG` reduces tree depth
- **LRU cache:** Helps when iteration patterns are spatially local; adds overhead for random access
- **Incremental vs random insertion:** Spatially-local (incremental) insertions are faster due to cache locality

## Project Structure

```
├── k2tree.go            # Core K2Tree (Go)
├── k2tree_grow.go       # Tree growth/insertion logic (Go)
├── iterator.go          # Row/column iterators (Go)
├── bitarray.go          # BitArray interface (Go)
├── slicearray.go        # Simple vector-backed bit array (Go)
├── binarylruarray.go    # LRU popcount cache (Go)
├── quartileindex.go     # Quartile popcount cache (Go)
├── pagedbitarray.go     # Multi-level paged bit array (Go)
├── paged_slicearray.go  # Paged slice arrays (Go)
├── pagefile.go          # Memory-mapped file support (Go)
├── insert_four_*.go     # SIMD-optimized bit insertion (Go, amd64 assembly)
├── k2config.go          # Configuration definitions (Go)
├── rust/                # Rust implementation
│   ├── src/
│   │   ├── k2tree/      # Core K2Tree
│   │   ├── bitarray/    # BitArray backends
│   │   ├── tests/       # Unit tests
│   │   └── error.rs     # Error types
│   ├── benches/         # Criterion benchmarks
│   └── examples/        # Example programs
└── *_test.go            # Tests and benchmarks (Go)
```

## Testing

```bash
# Go tests
go test ./...

# Go fuzz testing (Go 1.18+)
go test -fuzz=FuzzIteratorEdges -fuzztime=30s

# Rust tests
cd rust
cargo test

# Rust + doctests
cargo test --doc
```

## Limitations

- **No edge deletion:** The K²-tree is designed for append-only workloads. Removing edges is not supported.
- **No bulk insert:** Edges must be added one at a time via `Add()`.
- **Rust port gaps:** `PagedBitarray`, `PagedSliceArray`, and `Pagefile` are not yet ported to Rust.
- **No SIMD for insertFourBits in Rust:** The Go version uses hand-written amd64 assembly; the Rust version uses a portable byte-by-byte loop.

## Background

The K²-tree for graph representation was introduced by [Brisaboa, Ladra, and Navarro (2009)](https://doi.org/10.1007/978-3-642-03784-9_3) ([PDF](https://users.dcc.uchile.cl/~gnavarro/ps/spire09.1.pdf)) as a space-efficient alternative to traditional graph representations. It is particularly well-suited for:

- Web graphs (billions of edges, very sparse)
- Social network analysis
- Any domain where the adjacency matrix is sparse and read-heavy

## License

MIT — see [LICENSE](LICENSE)
