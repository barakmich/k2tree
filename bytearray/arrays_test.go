package bytearray

import (
	"bufio"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
)

func testCompareBaseline(t *testing.T, vec_b ByteArray) {
	tv := insertTestVector()
	vec_a := NewSlice()
	for i, x := range tv {
		b := byte(i)
		vec_a.Insert(x, []byte{b, b})
		vec_b.Insert(x, []byte{b, b})
		if vec_a.Len() != vec_b.Len() {
			t.Fatalf("Different Lengths after %d: %d %d", i, vec_a.Len(), vec_b.Len())
		}
	}

	for i := 0; i < vec_a.Len(); i++ {

		if vec_a.Get(i) != vec_b.Get(i) {
			t.Fatalf("Mismatched byte at %d: ex %v, got %v", i, vec_a.Get(i), vec_b.Get(i))
		}
	}

	if vec_a.PopCount(0, vec_a.Len()) != vec_b.PopCount(0, vec_b.Len()) {
		t.Fatalf("Mismatched PopCount for Whole String")
	}
	for i := 0; i < 500; i++ {
		x := rand.Intn(vec_a.Len())
		y := rand.Intn(vec_a.Len())
		if x > y {
			y, x = x, y
		}
		a := vec_a.PopCount(x, y)
		b := vec_b.PopCount(x, y)
		if a != b {
			t.Fatalf("Mismatched sub popcnt: %d to %d: %d vs %d", x, y, a, b)
		}
	}
}

func TestCompareBaselinePaged(t *testing.T) {
	vec := NewPaged(128, 0.8, 0.3)
	testCompareBaseline(t, vec)
}

func TestCompareBaselineFront(t *testing.T) {
	vec := NewFrontSlice(1024)
	testCompareBaseline(t, vec)
}

func TestCompareBaselineInt16(t *testing.T) {
	vec := NewInt16Index(NewSlice())
	testCompareBaseline(t, vec)
}

var insertTestVectorCache []int = nil

func insertTestVector() []int {
	if insertTestVectorCache != nil {
		return insertTestVectorCache
	}

	f, err := os.Open("insert_test.txt")
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(f)
	for {
		s, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		i, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			panic(err)
		}
		insertTestVectorCache = append(insertTestVectorCache, i/8)
	}

	return insertTestVectorCache
}
