package k2tree

import (
	"fmt"
	"reflect"
	"runtime"
)

func max(i, j int) int {
	if i > j {
		return i
	}
	return j
}

func min(i, j int) int {
	if i < j {
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

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func assert(test bool, errstr string) {
	if !test {
		panic(errstr)
	}
}

type twosHistogram struct {
	buckets [65]int
}

func (th *twosHistogram) Add(n int) {
	if n == 0 || n == 1 {
		th.buckets[0] += 1
		return
	}
	count := 0
	for n > 0 {
		n = n >> 1
		count += 1
	}
	th.buckets[count] += 1
}

func (th twosHistogram) String() string {
	out := "\n"
	for i, x := range th.buckets {
		out += fmt.Sprintf("%d: %d\n", 1<<i, x)
	}
	return out
}
