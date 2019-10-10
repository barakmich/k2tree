package k2tree

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
