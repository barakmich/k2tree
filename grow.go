package k2tree

import "fmt"

func (k *K2Tree) necessaryLayer(i int) int {
	// Level 1
	n := k.lk.kPerLayer * k.tk.kPerLayer
	l := 1
	for n < i {
		n *= k.tk.kPerLayer
		l++
	}
	return l
}

func (k *K2Tree) offsetTForLayer(i int, j int, l int) int {
	spl := uint(l-1)*(k.tk.shiftPerLayer) + k.lk.shiftPerLayer
	x := (i & (k.tk.maskPerLayer << spl)) >> spl
	y := (j & (k.tk.maskPerLayer << spl)) >> spl
	return (x * k.tk.kPerLayer) + y
}

func (k *K2Tree) offsetL(i int, j int) int {
	return ((i & k.lk.maskPerLayer) * k.lk.kPerLayer) + (j & k.lk.maskPerLayer)
}

func (k *K2Tree) growTree(i int) error {
	//n := k.necessaryLayer(i)
	panic("implement")
}

func (k *K2Tree) initTree(i, j int) error {
	l := k.necessaryLayer(max(i, j))
	k.t.Insert(k.tk.bitsPerLayer, 0)
	k.levels = l
	k.levelOffsets = make([]int, l+1)
	for x := l - 1; x > 0; x-- {
		k.levelOffsets[x] = k.tk.bitsPerLayer
	}
	return nil
}

func (k *K2Tree) insertToLayer(l int, layerCount int) error {
	if l == 0 {
		k.l.Insert(k.lk.bitsPerLayer, layerCount*k.lk.bitsPerLayer)
		return nil
	}
	k.t.Insert(k.tk.bitsPerLayer, (layerCount*k.tk.bitsPerLayer)+k.levelOffsets[l])
	for x := l - 1; x > 0; x-- {
		k.levelOffsets[x] += k.tk.bitsPerLayer
	}
	return nil
}

func (k *K2Tree) add(i, j int) error {
	fmt.Println("*****")
	level := k.levels
	if k.levelOffsets[level] != 0 {
		panic("top level is not offset 0?")
	}
	var levelOffset int
	var count int
	for level != 0 {
		fmt.Println("offset", k.levelOffsets)
		levelStart := k.levelOffsets[level]
		offset := k.offsetTForLayer(i, j, level)
		bitoff := levelStart + levelOffset + offset
		count = k.t.Count(levelStart, bitoff)
		if k.t.Get(bitoff) {
			levelOffset = count * k.tk.bitsPerLayer
		} else {
			k.t.Set(bitoff, true)
			k.insertToLayer(level-1, count)
			levelOffset = count * k.tk.bitsPerLayer
		}
		level--
	}
	offset := k.offsetL(i, j)
	bitoff := (count * k.lk.bitsPerLayer) + offset
	fmt.Println("bitoff:", bitoff)
	k.l.Set(bitoff, true)
	fmt.Println("*****")
	return nil
}

func (k *K2Tree) debug() string {
	s := fmt.Sprintln("T: ", k.t.debug())
	s += fmt.Sprintln("L: ", k.l.debug())
	s += fmt.Sprintln("Offset: ", k.levelOffsets)
	s += fmt.Sprintln("Levels: ", k.levels)
	return s
}
