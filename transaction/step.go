package transaction

import (
	"github.com/gluster/glusterd2/context"
	"github.com/gluster/glusterd2/utils"
)

// StepFunc is the function that is supposed to be run during a transaction step
type StepFunc func(*context.Context) error

const (
	//Leader is a constant string representing the leader node
	Leader = "leader"
	//All is a contant string representing all the nodes in a transaction
	All = "all"
)

// Step is a combination of a StepFunc and a list of nodes the step is supposed to be run on
//
// DoFunc performs does the action
// UndoFunc undoes anything done by DoFunc
type Step struct {
	DoFunc   StepFunc
	UndoFunc StepFunc
	Nodes    []string
}

// do runs the DoFunc on the nodes
func (s *Step) do(c *context.Context) error {
	for range s.Nodes {
		// TODO : Implement go routine
		// RunStepFunconNode(commandName, stepIndex, rollBack (false), c, n)
	}
	c.Log.WithField("step", utils.GetFuncName(s.DoFunc)).Debug("running step")

	return s.DoFunc(c)
}

// undo runs the UndoFunc on the nodes
func (s *Step) undo(c *context.Context) error {
	if s.UndoFunc != nil {
		for range s.Nodes {
			// TODO : Implement go routine
			// RunStepFunconNode(commandName, stepIndex, rollBack (false), c, n)
		}
		c.Log.WithField("undostep", utils.GetFuncName(s.DoFunc)).Debug("running undostep")
		return s.UndoFunc(c)
	}
	return nil
}
