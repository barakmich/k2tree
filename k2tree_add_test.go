package k2tree

import (
	"fmt"
	"testing"
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
		//fmt.Println("adding", x, x)
		kk.Add(x, x)
		//fmt.Println(kk.debug())
	}
	for x := 7; x >= 0; x-- {
		//fmt.Println("adding", x, x)
		k.Add(x, x)
		//fmt.Println(k.debug())
	}
	if k.t.Len() != kk.t.Len() {
		t.Error("lengths don't match in T")
	}
	for i := 0; i < k.t.Len(); i++ {
		if k.t.Get(i) != kk.t.Get(i) {
			t.Errorf("index %d doesn't match in T", i)
		}
	}
	if k.l.Len() != kk.l.Len() {
		t.Error("lengths don't match in L")
	}
	for i := 0; i < k.l.Len(); i++ {
		if k.l.Get(i) != kk.l.Get(i) {
			t.Errorf("index %d doesn't match in L", i)
		}
	}
}

func TestSixteenBPL(t *testing.T) {
	kk, err := New()
	if err != nil {
		t.Fatal(err)
	}
	kk.tk = sixteenBitsPerLayer
	kk.lk = fourBitsPerLayer
	base := 5000000
	for x := 0; x < 10000000; x++ {
		//fmt.Println("adding", x, x)
		kk.Add(base+x, base+x)
		if x%100000 == 0 {
			fmt.Println(x)
		}
	}
	//	fmt.Println(kk.debug())

	fmt.Printf("%#v\n", kk.Stats())

}
