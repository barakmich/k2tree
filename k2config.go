package k2tree

type Config struct {
	TreeLayerDef LayerDef
	CellLayerDef LayerDef
}

var DefaultConfig Config = SixteenFourConfig

var FourFourConfig Config = Config{
	TreeLayerDef: FourBitsPerLayer,
	CellLayerDef: FourBitsPerLayer,
}

var SixteenFourConfig Config = Config{
	TreeLayerDef: SixteenBitsPerLayer,
	CellLayerDef: FourBitsPerLayer,
}

var SixteenSixteenConfig Config = Config{
	TreeLayerDef: SixteenBitsPerLayer,
	CellLayerDef: SixteenBitsPerLayer,
}

var SixtySixteenConfig Config = Config{
	TreeLayerDef: SixtyFourBitsPerLayer,
	CellLayerDef: SixteenBitsPerLayer,
}
