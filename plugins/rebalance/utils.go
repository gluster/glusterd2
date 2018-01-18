package rebalance

import "time"

// TimeVal represents the time value
type TIMEVAL struct {
	TV_sec  uint64
	TV_usec uint64
}

var (
	hash uint64
)

func glusterdVolinfoResetStats(r RebalanceInfo) {
	r.RebalanceFiles = 0
	r.RebalanceData = 0
	r.LookedupFiles = 0
	r.RebalanceFailures = 0
	r.RebalanceTime = 0
	r.SkippedFiles = 0
}

func setCommitHash(r *RebalanceInfo) {
	tv := new(TIMEVAL)
	t := time.Now()
	tv.TV_sec = uint64(t.Day() * t.Hour() * t.Minute() * t.Second())
	tv.TV_usec = uint64((tv.TV_sec * 1000000))
	hash = tv.TV_sec << 3
	hash |= 1 << ((tv.TV_usec >> 10) % 3)
	r.CommitHash = hash
	return
}
