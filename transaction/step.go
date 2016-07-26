package transaction

import (
	"errors"

	"github.com/gluster/glusterd2/context"

	"github.com/pborman/uuid"
)

// StepFunc is the function that is supposed to be run during a transaction step
type StepFunc func(*context.Context) error

//const (
////Leader is a constant string representing the leader node
//Leader = "leader"
////All is a contant string representing all the nodes in a transaction
//All = "all"
//)
// XXX: Because Nodes are now uuid.UUID, string constants cannot be used in node lists
// TODO: Figure out an alternate method and re-enable. Or just remove it.

// Step is a combination of a StepFunc and a list of nodes the step is supposed to be run on
//
// DoFunc and UndoFunc are names of StepFuncs registered in the registry
// DoFunc performs does the action
// UndoFunc undoes anything done by DoFunc
type Step struct {
	DoFunc   string
	UndoFunc string
	Nodes    []uuid.UUID
}

// do runs the DoFunc on the nodes
func (s *Step) do(c *context.Context) error {
	for range s.Nodes {
		// RunStepFunconNode(s.DoFunc, n)
	}
	c.Log.WithField("step", s.DoFunc).Debug("running step")

	doFunc, ok := GetStepFunc(s.DoFunc)
	if !ok {
		return errors.New("StepFunc was not found")
	}
	return doFunc(c)
}

// undo runs the UndoFunc on the nodes
func (s *Step) undo(c *context.Context) error {
	if s.UndoFunc != "" {
		for range s.Nodes {
			// RunStepFunconNode(s.UndoFunc, n)
		}
		c.Log.WithField("undostep", s.UndoFunc).Debug("running undostep")

		undoFunc, ok := GetStepFunc(s.UndoFunc)
		if !ok {
			return errors.New("StepFunc was not found")
		}
		return undoFunc(c)
	}
	return nil
}
