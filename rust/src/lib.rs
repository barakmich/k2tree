//! # K2Tree
//!
//! A Rust implementation of the K2Tree data structure for compressed graph representation.
//!
//! K2Tree is a compressed data structure for representing large sparse adjacency matrices
//! (graphs). It uses a hierarchical tree structure with bit arrays to efficiently store
//! and query graph edges.
//!
//! ## Example
//!
//! ```
//! use k2tree::{K2Tree, SliceArray};
//!
//! let mut tree = K2Tree::new(SliceArray::new(), SliceArray::new());
//!
//! // Add edges
//! tree.add(0, 1).unwrap();
//! tree.add(0, 2).unwrap();
//! tree.add(1, 2).unwrap();
//!
//! // Query outgoing edges from node 0
//! let edges = tree.from(0).extract_all();
//! assert_eq!(edges.len(), 2);
//! ```

// Module declarations
mod bitarray;
mod error;
mod k2tree;

#[cfg(test)]
mod tests;

// Public API exports
pub use bitarray::{BitArray, LruArray, QuartileIndex, SliceArray};
pub use error::K2TreeError;
pub use k2tree::{
    Config, K2Tree, K2TreeIterator, LayerDef, Stats, DEFAULT_CONFIG, FOUR_BITS_PER_LAYER,
    FOUR_FOUR_CONFIG, SIXTEEN_BITS_PER_LAYER, SIXTEEN_FOUR_CONFIG, SIXTEEN_SIXTEEN_CONFIG,
};
