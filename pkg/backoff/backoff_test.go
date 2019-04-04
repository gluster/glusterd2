package backoff

import (
	"testing"
	"time"
)

var backOfftests = []struct {
	b    *BackOff
	want []time.Duration
}{
	{&BackOff{Factor: 2, Duration: time.Second}, []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second}},
	{&BackOff{Factor: 4, Duration: time.Second, MaxDuration: time.Second * 20}, []time.Duration{1 * time.Second, 4 * time.Second, 16 * time.Second, 20 * time.Second, 20 * time.Second}},
}

var backOffWithJittertests = []struct {
	b    *BackOff
	want []time.Duration
}{
	{&BackOff{Factor: 2, Duration: time.Second, JitterFactor: 2}, []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second}},
}

func TestBackOff_NextDuration(t *testing.T) {
	for _, backOffTest := range backOfftests {
		for _, dur := range backOffTest.want {
			got := backOffTest.b.NextDuration()
			if got != dur {
				t.Errorf("got: %v, want: %v", got, dur)
			}
		}
		if got := backOffTest.b.Attempts(); got != len(backOffTest.want) {
			t.Errorf("got :%d, want: %d", got, len(backOffTest.want))
		}
	}
}

func TestBackOff_Jitter(t *testing.T) {
	for _, backOffTest := range backOffWithJittertests {
		for _, dur := range backOffTest.want {
			got := backOffTest.b.NextDuration()
			max := (float64(dur) + float64(dur)*backOffTest.b.JitterFactor)
			if float64(got) > max {
				t.Errorf("got: %f, want < %f", float64(got), max)
			}
		}
	}
}
