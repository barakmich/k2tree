package k2tree

import (
	"fmt"
	"strings"
)

// K2Tree is the main data structure for this package. It represents a compressed representation of
// a graph adjacency matrix.
type K2Tree struct {
	t          bitarray
	l          bitarray
	tk         layerDef
	lk         layerDef
	count      int
	levels     int
	levelInfos levelInfos
}

type levelInfo struct {
	offset       int
	total        int
	midpoint     int
	fullPopCount int
	midPopCount  int
}

type levelInfos []levelInfo

func (li levelInfos) String() string {
	s := make([]string, len(li))
	for i, x := range li {
		s[i] = fmt.Sprintf("%d: Off: %d, Total %d, Midpoint %d, Pop: %d, MidPop: %d",
			i, x.offset, x.total, x.midpoint, x.fullPopCount, x.midPopCount)
	}
	return strings.Join(s, "\n")
}

// New creates a new K2 Tree with the default creation options.
func New() (*K2Tree, error) {
	return newK2Tree(func() bitarray {
		return newPagedSliceArray(100000)
	})
}

func newK2Tree(sliceFunc newBitArrayFunc) (*K2Tree, error) {
	t := sliceFunc()
	l := sliceFunc()
	return &K2Tree{
		t:      t,
		l:      l,
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

func intPow(a, b int) int {
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
func (k *K2Tree) Add(i, j int) error {
	if k.t.Len() == 0 {
		k.initTree(i, j)
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
	c := k.l.Total()
	bytes := k.l.Len() + k.t.Len()
	return Stats{
		BitsPerLink: float64(bytes) / float64(c),
		Links:       c,
		LevelInfo:   k.levelInfos,
		Bytes:       bytes >> 3,
	}
}

// Stats describes the memory usage of the K2 tree.
type Stats struct {
	BitsPerLink float64
	Links       int
	LevelInfo   levelInfos
	Bytes       int
}
