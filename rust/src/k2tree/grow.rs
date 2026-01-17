use super::K2Tree;
use crate::bitarray::BitArray;
use crate::error::K2TreeError;

/// Initializes a tree of the appropriate size to represent the given size.
pub fn init_tree<T: BitArray>(tree: &mut K2Tree<T>, size: usize) -> Result<(), K2TreeError> {
    let l = tree.necessary_layer(size);
    tree.tbits.insert(tree.tk.bits_per_layer, 0)?;
    tree.levels = l;
    tree.level_offsets = vec![0; l + 1];
    for x in (1..l).rev() {
        tree.level_offsets[x] = tree.tk.bits_per_layer;
    }
    Ok(())
}

/// Grows the K2Tree to be large enough to represent size.
pub fn grow_tree<T: BitArray>(tree: &mut K2Tree<T>, size: usize) -> Result<(), K2TreeError> {
    let n = tree.necessary_layer(size);
    while tree.levels != n {
        tree.tbits.insert(tree.tk.bits_per_layer, 0)?;
        tree.tbits.set(0, true);
        for x in (1..tree.level_offsets.len()).rev() {
            tree.level_offsets[x] += tree.tk.bits_per_layer;
        }
        tree.level_offsets.push(0);
        tree.levels += 1;
    }
    Ok(())
}

/// Inserts a new layer of bits in layer l given an offset determined by the above layer.
pub fn insert_to_layer<T: BitArray>(
    tree: &mut K2Tree<T>,
    l: usize,
    layer_count: usize,
) -> Result<(), K2TreeError> {
    if l == 0 {
        return tree
            .lbits
            .insert(tree.lk.bits_per_layer, layer_count * tree.lk.bits_per_layer);
    }

    let target_bit = layer_count * tree.tk.bits_per_layer;
    tree.tbits
        .insert(tree.tk.bits_per_layer, target_bit + tree.level_offsets[l])?;

    for x in (1..l).rev() {
        tree.level_offsets[x] += tree.tk.bits_per_layer;
    }
    Ok(())
}
