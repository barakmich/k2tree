package k2tree

type Config struct {
	TreeLayerDef LayerDef
	CellLayerDef LayerDef
}

var DefaultConfig Config = Config{
	TreeLayerDef: FourBitsPerLayer,
	CellLayerDef: FourBitsPerLayer,
}
