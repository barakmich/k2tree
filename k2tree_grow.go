package k2tree

import "fmt"

/*
Example 8x8 tree with 4 bits per Layer:

Matrix (i across j down):

01010000
10110001
01011010
01010100
10101000
01110101
10110001
10000000

     2           1
T: 1111 | 1111 0111 1111 1101
L  0110|0111|0101|0101|0001|1001|1000|1001|1011|1010|1100|1001|0001|0100

*/

// necessaryLayer computes the number of layers necessary to
// represent index i
func (k *K2Tree) necessaryLayer(i int) int {
	// Level 1
	n := k.lk.kPerLayer * k.tk.kPerLayer
	l := 1
	for n <= i {
		n *= k.tk.kPerLayer
		l++
	}
	return l
}

// offsetTForLayer returns the offset of i, j in layer l.
// In the above example, i=6, j=2:
// if l=2 == 1 (top right), if l=1 == 3 (bottom right)
func (k *K2Tree) offsetTForLayer(i int, j int, l int) int {
	spl := uint(l-1)*(k.tk.shiftPerLayer) + k.lk.shiftPerLayer
	x := (i & (k.tk.maskPerLayer << spl)) >> spl
	y := (j & (k.tk.maskPerLayer << spl)) >> spl
	return (x * k.tk.kPerLayer) + y
}

// Increment n by the amt
func (k *K2Tree) incrementNForLevel(n int, amt int, l int) int {
	spl := uint(l-1)*(k.tk.shiftPerLayer) + k.lk.shiftPerLayer
	return ((n >> spl) + amt) << spl
}

// returns the suboffset within the index of the lower bit layer
func (k *K2Tree) offsetL(i int, j int) int {
	return ((i & k.lk.maskPerLayer) * k.lk.kPerLayer) + (j & k.lk.maskPerLayer)
}

// growTree grows the K2Tree to be large enough to represent size
func (k *K2Tree) growTree(size int) error {
	n := k.necessaryLayer(size)
	for k.levels != n {
		err := k.tbits.Insert(k.tk.bitsPerLayer, 0)
		if err != nil {
			return err
		}
		k.tbits.Set(0, true)
		for x := len(k.levelOffsets) - 1; x > 0; x-- {
			k.levelOffsets[x] += k.tk.bitsPerLayer
		}
		k.levelOffsets = append(k.levelOffsets, 0)
		k.levels++
	}
	return nil
}

// initTree initializes a tree of the appropriate size
func (k *K2Tree) initTree(size int) error {
	l := k.necessaryLayer(size)
	err := k.tbits.Insert(k.tk.bitsPerLayer, 0)
	if err != nil {
		return err
	}
	k.levels = l
	k.levelOffsets = make([]int, l+1)
	for x := l - 1; x > 0; x-- {
		k.levelOffsets[x] = k.tk.bitsPerLayer
	}
	return nil
}

// insertToLayer inserts a new layersize of bits in layer l
// given an offset determined by the above layer.
func (k *K2Tree) insertToLayer(l int, layerCount int) error {
	if l == 0 {
		return k.lbits.Insert(k.lk.bitsPerLayer, layerCount*k.lk.bitsPerLayer)
	}
	targetBit := layerCount * k.tk.bitsPerLayer
	err := k.tbits.Insert(k.tk.bitsPerLayer, targetBit+k.levelOffsets[l])
	if err != nil {
		return err
	}
	for x := l - 1; x > 0; x-- {
		k.levelOffsets[x] += k.tk.bitsPerLayer
	}
	return nil
}

// add is the internal helper to set the appropriate bit at i,j.
func (k *K2Tree) add(i, j int) error {
	level := k.levels
	if k.levelOffsets[level] != 0 {
		panic("top level is not offset 0?")
	}
	var levelOffset int
	var count int
	for level != 0 {
		levelStart := k.levelOffsets[level]
		offset := k.offsetTForLayer(i, j, level)
		bitoff := levelStart + levelOffset + offset
		count = k.tbits.Count(levelStart, bitoff)
		//fmt.Println(
		//"level", level,
		//"levelStart", levelStart,
		//"offset", offset,
		//"bitoff", bitoff,
		//"count", count,
		//)
		if k.tbits.Get(bitoff) {
			levelOffset = count * k.tk.bitsPerLayer
		} else {
			k.tbits.Set(bitoff, true)
			k.insertToLayer(level-1, count)
			levelOffset = count * k.tk.bitsPerLayer
		}
		level--
	}
	offset := k.offsetL(i, j)
	bitoff := (count * k.lk.bitsPerLayer) + offset
	k.lbits.Set(bitoff, true)
	return nil
}

// debug debug-prints a K2Tree
func (k *K2Tree) debug() string {
	s := fmt.Sprintln("T: ", k.tbits.debug())
	s += fmt.Sprintln("L: ", k.lbits.debug())
	s += fmt.Sprintln("Offsets: ", k.levelOffsets, "Levels: ", k.levels)
	return s
}
