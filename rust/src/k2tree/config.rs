/// LayerDef defines the parameters for a layer in the K2Tree.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct LayerDef {
    pub bits_per_layer: usize,
    pub k_per_layer: usize,
    pub mask_per_layer: usize,
    pub shift_per_layer: usize,
}

/// FourBitsPerLayer is a layer definition with 4 bits per layer (k=2).
pub const FOUR_BITS_PER_LAYER: LayerDef = LayerDef {
    bits_per_layer: 4,
    k_per_layer: 2,
    mask_per_layer: 0x1,
    shift_per_layer: 1,
};

/// SixteenBitsPerLayer is a layer definition with 16 bits per layer (k=4).
pub const SIXTEEN_BITS_PER_LAYER: LayerDef = LayerDef {
    bits_per_layer: 16,
    k_per_layer: 4,
    mask_per_layer: 0x3,
    shift_per_layer: 2,
};

/// Config defines the configuration for a K2Tree.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Config {
    pub tree_layer_def: LayerDef,
    pub cell_layer_def: LayerDef,
}

/// FourFourConfig uses 4 bits per layer for both tree and cell layers.
pub const FOUR_FOUR_CONFIG: Config = Config {
    tree_layer_def: FOUR_BITS_PER_LAYER,
    cell_layer_def: FOUR_BITS_PER_LAYER,
};

/// SixteenFourConfig uses 16 bits per layer for tree layers and 4 bits for cell layers.
pub const SIXTEEN_FOUR_CONFIG: Config = Config {
    tree_layer_def: SIXTEEN_BITS_PER_LAYER,
    cell_layer_def: FOUR_BITS_PER_LAYER,
};

/// SixteenSixteenConfig uses 16 bits per layer for both tree and cell layers.
pub const SIXTEEN_SIXTEEN_CONFIG: Config = Config {
    tree_layer_def: SIXTEEN_BITS_PER_LAYER,
    cell_layer_def: SIXTEEN_BITS_PER_LAYER,
};

/// DefaultConfig is the default configuration for K2Trees (SixteenFourConfig).
pub const DEFAULT_CONFIG: Config = SIXTEEN_FOUR_CONFIG;
