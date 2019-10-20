package k2tree

type Config struct {
	TreeLayerDef LayerDef
	CellLayerDef LayerDef
}

type LayerDef struct {
	bitsPerLayer  int
	kPerLayer     int
	maskPerLayer  int
	shiftPerLayer uint
}

var FourBitsPerLayer = LayerDef{
	bitsPerLayer:  4,
	kPerLayer:     2,
	maskPerLayer:  0x1,
	shiftPerLayer: 1,
}

var SixteenBitsPerLayer = LayerDef{
	bitsPerLayer:  16,
	kPerLayer:     4,
	maskPerLayer:  0x3,
	shiftPerLayer: 2,
}

var SixtyFourBitsPerLayer = LayerDef{
	bitsPerLayer:  64,
	kPerLayer:     8,
	maskPerLayer:  0x7,
	shiftPerLayer: 3,
}

var TwoFiftySixBitsPerLayer = LayerDef{
	bitsPerLayer:  256,
	kPerLayer:     16,
	maskPerLayer:  0xf,
	shiftPerLayer: 4,
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
