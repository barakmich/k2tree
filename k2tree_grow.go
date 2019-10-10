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
		err := k.tbits.Insert(k.tk.bitsPerLayer, 0)
		if err != nil {
			return err
		}
		k.tbits.Set(0, true)
		for x := len(k.levelInfos) - 1; x > 0; x-- {
			k.levelInfos[x].offset += k.tk.bitsPerLayer
		}
		k.levelInfos = append(k.levelInfos, levelInfo{
			offset:       0,
			total:        k.tk.bitsPerLayer,
			midpoint:     k.tk.bitsPerLayer >> 1,
			fullPopCount: 1,
		})
		k.levels++
	}
	return nil
}

func (k *K2Tree) initTree(i, j int) error {
	l := k.necessaryLayer(max(i, j))
	err := k.tbits.Insert(k.tk.bitsPerLayer, 0)
	if err != nil {
		return err
	}
	k.levels = l
	k.levelInfos = make([]levelInfo, l+1)
	k.levelInfos[l] = levelInfo{
		offset:       0,
		total:        k.tk.bitsPerLayer,
		midpoint:     k.tk.bitsPerLayer >> 1,
		fullPopCount: 0,
	}
	for x := l - 1; x > 0; x-- {
		k.levelInfos[x].offset = k.tk.bitsPerLayer
	}
	return nil
}

func (k *K2Tree) insertToLayer(l int, layerCount int) error {
	if l == 0 {
		return k.lbits.Insert(k.lk.bitsPerLayer, layerCount*k.lk.bitsPerLayer)
	}
	targetBit := layerCount * k.tk.bitsPerLayer
	err := k.tbits.Insert(k.tk.bitsPerLayer, targetBit+k.levelInfos[l].offset)
	if err != nil {
		return err
	}
	k.insertLevelInfo(l, targetBit)
	for x := l - 1; x > 0; x-- {
		k.levelInfos[x].offset += k.tk.bitsPerLayer
	}
	return nil
}

func (k *K2Tree) insertLevelInfo(level int, targetBit int) {
	li := k.levelInfos[level]
	li.total += k.tk.bitsPerLayer
	adjust := k.tk.bitsPerLayer >> 1
	if targetBit <= li.midpoint {
		li.midpoint += k.tk.bitsPerLayer
		c := k.tbits.Count(li.midpoint-adjust, li.midpoint)
		li.midpoint -= adjust
		li.midPopCount -= c

	} else {
		c := k.tbits.Count(li.midpoint, li.midpoint+adjust)
		li.midpoint += adjust
		li.midPopCount += c
	}
	k.levelInfos[level] = li
}

func (k *K2Tree) countLevelStartHelper(level int, levelOffset int, subindex int) int {
	li := k.levelInfos[level]
	levelStart := li.offset
	bitoff := levelStart + levelOffset + subindex
	if bitoff > li.midpoint+levelStart {
		if bitoff-li.midpoint < li.total-bitoff {
			c := k.tbits.Count(li.midpoint+levelStart, bitoff)
			return li.midPopCount + c
		} else {
			c := k.tbits.Count(bitoff, levelStart+li.total)
			return li.fullPopCount - c
		}
	}
	//Fallthrough
	count := k.tbits.Count(levelStart, bitoff)
	return count
}

func (k *K2Tree) setHelper(bitoff int, level int, value bool) {
	k.levelInfos[level].fullPopCount++
	levelStart := k.levelInfos[level].offset
	if bitoff-levelStart < k.levelInfos[level].midpoint {
		k.levelInfos[level].midPopCount++
	}
	k.tbits.Set(bitoff, value)
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
		count = k.countLevelStartHelper(level, levelOffset, offset)
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
			k.setHelper(bitoff, level, true)
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

func (k *K2Tree) debug() string {
	s := fmt.Sprintln("T: ", k.tbits.debug())
	s += fmt.Sprintln("L: ", k.lbits.debug())
	s += fmt.Sprintln("Offsets: ", k.levelInfos, "Levels: ", k.levels)
	return s
}
