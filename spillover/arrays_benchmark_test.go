package spillover

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
)

func BenchmarkInsertPatternArray(b *testing.B) {
	for _, arraytype := range arrayTypes {
		b.Run(fmt.Sprintf(arraytype.name), func(b *testing.B) {
			tv := insertTestVector()
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				vec := arraytype.makeArray()
				for i := 0; i < 4; i++ {
					for _, x := range tv {
						vec.Insert(x, []byte{0x01, 0x01})
					}
				}
			}

		})
	}
}

type testarraytype struct {
	makeArray func() testArray
	name      string
}

var arrayTypes []testarraytype = []testarraytype{
	{
		makeArray: func() testArray {
			return &sliceTest{}
		},
		name: "Slice:::::::::",
	},
	{
		makeArray: func() testArray {
			return newFrontSliceTest(1024)
		},
		name: "FrontSlice1k:::",
	},
	{
		makeArray: func() testArray {
			return newFrontSliceTest(1024 * 1024)
		},
		name: "FrontSlice1M:::",
	},
	{
		makeArray: func() testArray {
			return New(512, 0.8, 0.5)
		},
		name: "Spillover:512:80:50",
	},
	{
		makeArray: func() testArray {
			return New(512, 0.8, 0.3)
		},
		name: "Spillover:512:80:30",
	},
	{
		makeArray: func() testArray {
			return New(512, 0.9, 0.7)
		},
		name: "Spillover:512:90:70",
	},
	{
		makeArray: func() testArray {
			return New(512, 0.9, 0.5)
		},
		name: "Spillover:512:90:50",
	},
	{
		makeArray: func() testArray {
			return New(512, 0.9, 0.3)
		},
		name: "Spillover:512:90:30",
	},
	{
		makeArray: func() testArray {
			return New(1024, 0.8, 0.5)
		},
		name: "Spillover:1024:80:50",
	},
	{
		makeArray: func() testArray {
			return New(2048, 0.8, 0.5)
		},
		name: "Spillover:2048:80:50",
	},
}

func TestCompareBaselineFront(t *testing.T) {
	tv := insertTestVector()
	vec_a := newSliceTest()
	vec_b := newFrontSliceTest(1024)
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
