package logging

import (
	"fmt"
	"path"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	// SourceField is the field name used for logging source location.
	SourceField = "source"
	gd2Repo     = "github.com/gluster/glusterd2"
)

// SourceLocationHook is a type that implements the logrus.Hook interface.
type SourceLocationHook struct{}

// Levels returns all logrus levels. The hook is fired only for those log
// levels returned by this function.
func (hook SourceLocationHook) Levels() []logrus.Level {
	// TODO: Can optionally make this specific to Debug level.
	return logrus.AllLevels
}

// Fire adds file name, function name and line number to the log entry.
func (hook SourceLocationHook) Fire(entry *logrus.Entry) error {
	pcs := make([]uintptr, 3)
	n := runtime.Callers(6, pcs)
	if n == 0 {
		return nil
	}

	frames := runtime.CallersFrames(pcs)
	for {
		frame, more := frames.Next()
		if strings.Contains(frame.File, gd2Repo) && !strings.Contains(frame.File, "vendor") {
			entry.Data[SourceField] = fmt.Sprintf("[%s:%d:%s]", path.Base(frame.File), frame.Line, path.Base(frame.Function))
			break
		}
		if !more {
			break
		}
	}

	return nil
}
