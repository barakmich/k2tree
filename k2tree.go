package k2tree

// K2Tree is the main data structure for this package. It represents a compressed representation of
// a graph adjacency matrix.
type K2Tree struct {
	t            bitarray
	l            bitarray
	tk           layerDef
	lk           layerDef
	count        int
	levels       int
	levelOffsets []int
}

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
	maskPerLayer:  0x2,
	shiftPerLayer: 2,
}

// New creates a new K2Tree.
func New() (*K2Tree, error) {
	t := &sliceArray{}
	return &K2Tree{
		t:      t,
		l:      &sliceArray{},
		tk:     fourBitsPerLayer,
		lk:     fourBitsPerLayer,
		levels: 0,
	}, nil
}

func max(i, j int) int {
	if i > j {
		return i
	}
	return j
}

func iPow(a, b int) int {
	var result = 1

	for 0 != b {
		if 0 != (b & 1) {
			result *= a
		}
		b >>= 1
		a *= a
	}

	return result
}

func (k *K2Tree) max() int {
	if k.levels == 0 {
		return 0
	}
	x := iPow(k.tk.kPerLayer, k.levels) * k.lk.kPerLayer
	return x
}

// Add asserts the existence of a link from node i to node j.
func (k *K2Tree) Add(i, j int) error {
	if k.t.Len() == 0 {
		k.initTree(i, j)
	} else if i > k.max() || j > k.max() {
		err := k.growTree(max(i, j))
		if err != nil {
			return err
		}
	}
	return k.add(i, j)
}

// Stats returns some statistics about the memory usage of the K2 tree.
func (k *K2Tree) Stats() Stats {
	c := k.l.Count(0, k.l.Len())
	return Stats{
		BitsPerLink: float64(k.t.Len()+k.l.Len()) / float64(c),
		Links:       c,
	}
}

// Stats describes the memory usage of the K2 tree.
type Stats struct {
	BitsPerLink float64
	Links       int
}