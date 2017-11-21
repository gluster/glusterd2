package georeplication

import "fmt"

type actionType uint16

const (
	actionCreate actionType = iota
	actionStart
	actionStop
	actionPause
	actionResume
	actionDelete
)

func (i actionType) String() string {
	switch i {
	case actionCreate:
		return "create"
	case actionStart:
		return "start"
	case actionStop:
		return "stop"
	case actionPause:
		return "pause"
	case actionResume:
		return "resume"
	case actionDelete:
		return "delete"
	default:
		return fmt.Sprintf("actionType(%d)", i)
	}
}
