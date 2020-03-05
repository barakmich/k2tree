package bytearray

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

func BenchmarkSpilloverMatrix(b *testing.B) {
	highwaters := []float64{0.75, 0.8, 0.9}
	lowwaters := []float64{0.3, 0.5, 0.7}
	multipliers := []bool{false, true}
	pagesizes := []int{1024, 4 * 1024}
	for _, p := range pagesizes {
		for _, m := range multipliers {
			for _, h := range highwaters {
				for _, l := range lowwaters {
					b.Run(fmt.Sprintf("%d:%.2f:%.2f:%v", p, h, l, m), func(b *testing.B) {
						tv := insertTestVector()
						b.ResetTimer()
						for n := 0; n < b.N; n++ {
							vec := NewSpillover(p, h, l, m)
							for i := 0; i < 10; i++ {
								for _, x := range tv {
									vec.Insert(x, []byte{0x01, 0x01})
								}
							}
						}

					})
				}
			}
		}
	}
}

type testarraytype struct {
	makeArray func() ByteArray
	name      string
}

var arrayTypes []testarraytype = []testarraytype{
	{
		makeArray: func() ByteArray {
			return &SliceArray{}
		},
		name: "Slice:::::::::",
	},
	{
		makeArray: func() ByteArray {
			return NewFrontSlice(1024)
		},
		name: "FrontSlice1k:::",
	},
	{
		makeArray: func() ByteArray {
			return NewFrontSlice(1024 * 1024)
		},
		name: "FrontSlice1M:::",
	},
	{
		makeArray: func() ByteArray {
			return NewSpillover(1024, 0.8, 0.5, true)
		},
		name: "Spillover::1024:80:50:2x",
	},
	{
		makeArray: func() ByteArray {
			return NewSpillover(4096, 0.8, 0.3, false)
		},
		name: "Spillover::1024:80:30:1x",
	},
}

func TestCompareBaselineFront(t *testing.T) {
	tv := insertTestVector()
	vec_a := NewSlice()
	vec_b := NewFrontSlice(1024)
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
