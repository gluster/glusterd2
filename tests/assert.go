// Package tests borrows Assert() from https://github.com/heketi/heketi
package tests

import (
	"runtime"
	"testing"
)

// Assert provides a simple assert call for unit and functional tests
func Assert(t *testing.T, b bool) {
	if !b {
		pc, file, line, _ := runtime.Caller(1)
		callFuncInfo := runtime.FuncForPC(pc)

		t.Errorf("\n\rASSERT:\tfunc (%s) 0x%x\n\r\tFile %s:%d",
			callFuncInfo.Name(),
			pc,
			file,
			line)
		t.FailNow()
	}
}
