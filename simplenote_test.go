package simplenote

import (
	"testing"
	"time"
)

func TestTimeParse(t *testing.T) {
	t1 := time.Now()
	st := timeToStr(t1)
	t2 := strToTime(st)
	st2 := timeToStr(t2)
	if st != st2 {
		t.Fatalf("%s != %s", st, st2)
	}
	diff := int64(t1.Sub(t2))
	if diff < 0 {
		diff = -diff
	}
	// in nanoseconds
	if diff > 1000 {
		t.Fatalf("diff (%d) is > 1000", diff)
	}
}
