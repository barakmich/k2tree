package k2tree

import (
	"fmt"
	"testing"
	"time"
)

func TestSimpleAdd(t *testing.T) {
	k, err := New()
	if err != nil {
		t.Fatal(err)
	}
	kk, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for x := 0; x < 8; x++ {
		kk.Add(x, x)
	}
	for x := 7; x >= 0; x-- {
		k.Add(x, x)
	}
	if k.tbits.Len() != kk.tbits.Len() {
		t.Error("lengths don't match in T")
	}
	for i := 0; i < k.tbits.Len(); i++ {
		if k.tbits.Get(i) != kk.tbits.Get(i) {
			t.Errorf("index %d doesn't match in T", i)
		}
	}
	if k.lbits.Len() != kk.lbits.Len() {
		t.Error("lengths don't match in L")
	}
	for i := 0; i < k.lbits.Len(); i++ {
		if k.lbits.Get(i) != kk.lbits.Get(i) {
			t.Errorf("index %d doesn't match in L", i)
		}
	}
}

// TestSixteenBPL inserts a diagonal of edges, starting at a large offset.
func TestSixteenBPL(t *testing.T) {
	kk, err := NewWithConfig(
		Config{
			TreeLayerDef: SixteenBitsPerLayer,
			CellLayerDef: SixteenBitsPerLayer,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	base := 5000000
	td := time.Now()
	for x := 0; x < 1000000; x++ {
		kk.Add(base+x, base+x)
		if x%100000 == 0 {
			newt := time.Now()
			fmt.Println(x, newt.Sub(td))
			td = newt
		}
	}
	fmt.Println(kk.Stats())
}
