# K2Tree Benchmarks

This directory contains benchmarks for the K2Tree implementation using the basic SliceArray backend.

## Benchmarks

### Iterator Benchmarks

- **extract_20_slice**: Measures the performance of extracting all outgoing edges from a specific node (node 20) in a small preloaded tree. Tests iterator performance.

### Population Benchmarks

- **rand_pop_1k_slice**: Adds 1,000 random edges to an empty tree with SIXTEEN_SIXTEEN_CONFIG
- **inc_pop_1k_slice**: Adds 1,000 edges with incremental (spatially local) pattern with SIXTEEN_FOUR_CONFIG
- **rand_pop_10k_slice**: Adds 10,000 random edges to an empty tree with SIXTEEN_SIXTEEN_CONFIG
- **inc_pop_10k_slice**: Adds 10,000 edges with incremental pattern with SIXTEEN_FOUR_CONFIG

The "incremental" pattern simulates spatially local edge additions, which is common in real-world graph construction. The random pattern tests performance with uniform distribution across the ID space.

## Running Benchmarks

```bash
# Run all benchmarks
cargo bench --bench k2tree_bench

# Run a specific benchmark
cargo bench --bench k2tree_bench -- extract_20_slice

# Generate detailed reports
cargo bench --bench k2tree_bench -- --verbose
```

Benchmark results are saved to `target/criterion/` with detailed HTML reports.

## Notes

These benchmarks only use the basic `SliceArray` implementation. The Go version includes benchmarks for various optimized BitArray implementations (LRU-indexed, paged, etc.) which are not yet translated to Rust.
