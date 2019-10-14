package k2tree

import "fmt"

// K2Tree is the main data structure for this package. It represents a compressed representation of
// a graph adjacency matrix.
type K2Tree struct {
	tbits        bitarray
	lbits        bitarray
	tk           LayerDef
	lk           LayerDef
	count        int
	levels       int
	levelOffsets []int
}

// New creates a new K2 Tree with the default creation options.
func New() (*K2Tree, error) {
	return NewWithConfig(DefaultConfig)
}

func NewWithConfig(config Config) (*K2Tree, error) {
	return newK2Tree(func() bitarray {
		//return newQuartileIndex(newTraceArray(&sliceArray{}))
		//return newDebugArray(newInt16Index(newTraceArray(&sliceArray{})))
		//return newCompareArray(&sliceArray{}, newInt16Index(&sliceArray{}))
		return newInt16Index(newTraceArray(&sliceArray{}))
		//return newDebugArray(&sliceArray{})
	}, config)
}

func newK2Tree(sliceFunc newBitArrayFunc, config Config) (*K2Tree, error) {
	t := sliceFunc()
	l := &sliceArray{}
	return &K2Tree{
		tbits:  t,
		lbits:  l,
		tk:     config.TreeLayerDef,
		lk:     config.CellLayerDef,
		levels: 0,
	}, nil
}

// maxIndex returns the largest node index representable by this
// K2Tree.
func (k *K2Tree) maxIndex() int {
	if k.levels == 0 {
		return 0
	}
	x := intPow(k.tk.kPerLayer, k.levels) * k.lk.kPerLayer
	return x
}

// Add asserts the existence of a link from node i to node j.
// i and j are zero-indexed, the tree will grow to support them if larger
// than the tree.
func (k *K2Tree) Add(i, j int) error {
	if k.tbits.Len() == 0 {
		k.initTree(max(i, j))
	} else if i >= k.maxIndex() || j >= k.maxIndex() {
		err := k.growTree(max(i, j))
		if err != nil {
			return err
		}
	}
	return k.add(i, j)
}

// Stats returns some statistics about the memory usage of the K2 tree.
func (k *K2Tree) Stats() Stats {
	c := k.lbits.Total()
	bytes := k.lbits.Len() + k.tbits.Len()
	return Stats{
		BitsPerLink:  float64(bytes) / float64(c),
		Links:        c,
		LevelOffsets: k.levelOffsets,
		Bytes:        bytes >> 3,
		TDebug:       k.tbits.debug(),
		LDebug:       k.lbits.debug(),
	}
}

// Stats describes the memory usage of the K2 tree.
type Stats struct {
	BitsPerLink  float64
	Links        int
	LevelOffsets []int
	Bytes        int
	TDebug       string
	LDebug       string
}

func (st Stats) String() string {
	return fmt.Sprintf(`
Bits Per Link: %v
Links: %d
LevelOffsets: %v
Bytes: %d
TDebug: %s
LDebug: %s
	`,
		st.BitsPerLink,
		st.Links,
		st.LevelOffsets,
		st.Bytes,
		st.TDebug,
		st.LDebug,
	)
}
