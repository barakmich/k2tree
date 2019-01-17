package k2tree

func (k *K2Tree) necessaryLayer(i int) int {
	// Level 1
	n := k.lk.kPerLayer * k.tk.kPerLayer
	l := 1
	for i < n {
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
	return nil
}

func (k *K2Tree) initTree(i, j int) error {
	l := k.necessaryLayer(max(i, j))
	k.t.Insert(k.tk.bitsPerLayer, 0)
	k.levels = l
	return nil
}
