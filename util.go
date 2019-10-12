package k2tree

import (
	"reflect"
	"runtime"
)

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

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func assert(test bool, errstr string) {
	if test {
		panic(errstr)
	}
}
