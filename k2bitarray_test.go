package k2tree

import (
	"fmt"
	"testing"
)

func TestPopulate50k(t *testing.T) {
	for _, bitarray := range testBitArrayTypes {
		t.Run(fmt.Sprintf(bitarray.name), func(t *testing.T) { testPopulate(t, bitarray.create, 50000) })
	}
}

func testPopulate(t testing.TB, ba newBitArrayFunc, n int) *K2Tree {
	k2, err := newK2Tree(func() bitarray {
		x := ba()
		return newTraceArray(x)
	}, Config{
		TreeLayerDef: SixteenBitsPerLayer,
		CellLayerDef: SixteenBitsPerLayer,
	})
	if err != nil {
		t.Fatal(err)
	}
	populateRandomTree(n, n*2, k2)
	return k2
}

func BenchmarkPopulate50k(b *testing.B) {
	for _, bitarray := range testBitArrayTypes {
		b.Run(fmt.Sprintf(bitarray.name), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				k2 := testPopulate(b, bitarray.create, 50000)
				b.SetBytes(int64(k2.tbits.(*traceArray).data.CountLengths) / 8)
			}
		})
	}
}

func BenchmarkPopulate100k(b *testing.B) {
	for _, bitarray := range testBitArrayTypes {
		b.Run(fmt.Sprintf(bitarray.name), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				k2 := testPopulate(b, bitarray.create, 100000)
				b.SetBytes(int64(k2.tbits.(*traceArray).data.CountLengths) / 8)
			}
		})
	}
}

func TestRandAdd(t *testing.T) {
	k2, err := newK2Tree(
		func() bitarray {
			return newPagedSliceArray(10)
		},
		Config{
			TreeLayerDef: SixteenBitsPerLayer,
			CellLayerDef: FourBitsPerLayer,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	k2.Add(48081, 27887)
	k2.Add(31847, 34059)
	k2.Add(2081, 41318)
	k2.Add(4425, 22540)
	k2.Add(40456, 3300)
	tmp := k2.Row(2081).ExtractAll()
	if len(tmp) != 1 && tmp[0] != 41318 {
		t.Errorf("Unmatched 2081")
	}
	tmp = k2.Row(40456).ExtractAll()
	if len(tmp) != 1 && tmp[0] != 3300 {
		t.Errorf("Unmatched 40456")
	}

}
