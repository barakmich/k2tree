package k2tree

type layerDef struct {
	bitsPerLayer  int
	kPerLayer     int
	maskPerLayer  int
	shiftPerLayer uint
}

var fourBitsPerLayer = layerDef{
	bitsPerLayer:  4,
	kPerLayer:     2,
	maskPerLayer:  0x1,
	shiftPerLayer: 1,
}

var sixteenBitsPerLayer = layerDef{
	bitsPerLayer:  16,
	kPerLayer:     4,
	maskPerLayer:  0x3,
	shiftPerLayer: 2,
}

var sixtyFourBitsPerLayer = layerDef{
	bitsPerLayer:  64,
	kPerLayer:     8,
	maskPerLayer:  0x7,
	shiftPerLayer: 3,
}
