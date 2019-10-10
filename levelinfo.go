package k2tree

import (
	"fmt"
	"strings"
)

type levelInfo struct {
	offset       int
	total        int
	midpoint     int
	fullPopCount int
	midPopCount  int
}

type levelInfos []levelInfo

func (li levelInfos) String() string {
	s := make([]string, len(li))
	for i, x := range li {
		s[i] = fmt.Sprintf("%d: Off: %d, Total %d, Midpoint %d, Pop: %d, MidPop: %d",
			i, x.offset, x.total, x.midpoint, x.fullPopCount, x.midPopCount)
	}
	return strings.Join(s, "\n")
}
