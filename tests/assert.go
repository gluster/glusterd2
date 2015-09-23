// From https://github.com/heketi/heketi
package tests

import (
	"runtime"
	"testing"
)

// Simple assert call for unit and functional tests
func Assert(t *testing.T, b bool) {
	if !b {
		pc, file, line, _ := runtime.Caller(1)
		caller_func_info := runtime.FuncForPC(pc)

		t.Errorf("\n\rASSERT:\tfunc (%s) 0x%x\n\r\tFile %s:%d",
			caller_func_info.Name(),
			pc,
			file,
			line)
		t.FailNow()
	}
}
