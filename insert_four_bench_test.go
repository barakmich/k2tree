package k2tree

import (
	"fmt"
	"math/rand"
	"testing"
)

var insertFourSizes = []int{8, 15, 16, 31, 32, 63, 64, 256, 1024, 4096, 65536}

func benchmarkInsertFour(b *testing.B, size int, fn func([]byte, byte) byte) {
	rng := rand.New(rand.NewSource(1))
	buf := make([]byte, size)
	for i := range buf {
		buf[i] = byte(rng.Intn(256))
	}
	inByte := byte(0xA0)

	b.SetBytes(int64(size))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inByte = fn(buf, inByte)
	}
}

func BenchmarkInsertFourGo(b *testing.B) {
	for _, size := range insertFourSizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			benchmarkInsertFour(b, size, insertFourBitsGo)
		})
	}
}

func BenchmarkInsertFourHW(b *testing.B) {
	for _, size := range insertFourSizes {
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			benchmarkInsertFour(b, size, insertFourBits)
		})
	}
}
