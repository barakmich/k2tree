package k2tree

import "fmt"

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
	n := k.necessaryLayer(i)
	for k.levels != n {
		err := k.t.Insert(k.tk.bitsPerLayer, 0)
		if err != nil {
			return err
		}
		k.t.Set(0, true)
		for x := len(k.levelInfos) - 1; x > 0; x-- {
			k.levelInfos[x].offset += k.tk.bitsPerLayer
		}
		k.levelInfos = append(k.levelInfos, levelInfo{})
		k.levels++
	}
	return nil
}

func (k *K2Tree) initTree(i, j int) error {
	l := k.necessaryLayer(max(i, j))
	err := k.t.Insert(k.tk.bitsPerLayer, 0)
	if err != nil {
		return err
	}
	k.levels = l
	k.levelInfos = make([]levelInfo, l+1)
	for x := l - 1; x > 0; x-- {
		k.levelInfos[x].offset = k.tk.bitsPerLayer
	}
	return nil
}

func (k *K2Tree) insertToLayer(l int, layerCount int) error {
	if l == 0 {
		return k.l.Insert(k.lk.bitsPerLayer, layerCount*k.lk.bitsPerLayer)
	}
	err := k.t.Insert(k.tk.bitsPerLayer, (layerCount*k.tk.bitsPerLayer)+k.levelInfos[l].offset)
	if err != nil {
		return err
	}
	for x := l - 1; x > 0; x-- {
		k.levelInfos[x].offset += k.tk.bitsPerLayer
	}
	return nil
}

func (k *K2Tree) add(i, j int) error {
	level := k.levels
	if k.levelInfos[level].offset != 0 {
		panic("top level is not offset 0?")
	}
	var levelOffset int
	var count int
	for level != 0 {
		levelStart := k.levelInfos[level].offset
		offset := k.offsetTForLayer(i, j, level)
		bitoff := levelStart + levelOffset + offset
		count = k.t.Count(levelStart, bitoff)
		//fmt.Println(
		//"level", level,
		//"levelStart", levelStart,
		//"offset", offset,
		//"bitoff", bitoff,
		//"count", count,
		//)
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
	k.l.Set(bitoff, true)
	return nil
}

func (k *K2Tree) debug() string {
	s := fmt.Sprintln("T: ", k.t.debug())
	s += fmt.Sprintln("L: ", k.l.debug())
	s += fmt.Sprintln("Offsets: ", k.levelInfos, "Levels: ", k.levels)
	return s
}
