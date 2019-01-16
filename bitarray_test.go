package k2tree

import (
	"fmt"
	"testing"
)

func TestSmoke(t *testing.T) {
	s := &sliceArray{}
	s.Insert(24, 0)
	s.Set(3, true)
	fmt.Println(s.debug())
	if s.Count(0, 2) != 0 {
		t.Error("wrong count")
	}
	if s.Count(2, 3) != 0 {
		t.Error("end inclusive?")
	}
	if s.Count(2, 4) != 1 {
		t.Error("can't count?")
	}
	if !s.Get(3) {
		t.Error("can't retrieve?")
	}
	for x := 0; x < 24; x++ {
		s.Set(x, true)
	}
	fmt.Println(s.debug())
	if s.Count(0, 8) != 8 {
		t.Error("wrong count")
	}
	if s.Count(0, s.Len()) != 24 {
		t.Error("wrong count")
	}
}

func TestEasyInsert(t *testing.T) {
	s := &sliceArray{}
	s.Insert(24, 0)
	s.Set(3, true)
	s.Insert(8, 0)
	fmt.Println(s.debug())
	if s.Get(3) {
		t.Error("new 3 should not be set")
	}
	if !s.Get(11) {
		t.Error("new 11 should be set")
	}
}

func TestByteInsert(t *testing.T) {
	s := &sliceArray{}
	s.Insert(24, 0)
	s.Set(11, true)
	s.Set(6, true)
	s.Set(2, true)
	fmt.Println(s.debug())
	s.Insert(8, 4)
	fmt.Println(s.debug())
	if s.Get(11) {
		t.Error("new 11 should not be set")
	}
	if !s.Get(19) {
		t.Error("new 19 should be set")
	}
	if !s.Get(14) {
		t.Error("new 14 should be set")
	}
	if !s.Get(2) {
		t.Error("new 2 should be set")
	}

}

func TestNibbleInsert(t *testing.T) {
	s := &sliceArray{}
	s.Insert(24, 0)
	s.Set(11, true)
	s.Set(6, true)
	s.Set(2, true)
	fmt.Println(s.debug())
	s.Insert(4, 4)
	fmt.Println(s.debug())
	if s.Get(11) {
		t.Error("new 11 should not be set")
	}
	if !s.Get(15) {
		t.Error("new 15 should be set")
	}
	if !s.Get(10) {
		t.Error("new 10 should be set")
	}
	if !s.Get(2) {
		t.Error("new 2 should be set")
	}

}
