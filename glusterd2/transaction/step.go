package transaction

import (
	"errors"

	"github.com/gluster/glusterd2/glusterd2/gdctx"

	"github.com/pborman/uuid"
)

// StepFunc is the function that is supposed to be run during a transaction step
type StepFunc func(TxnCtx) error

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

var (
	// ErrStepFuncNotFound is returned if the stepfunc isn't found.
	ErrStepFuncNotFound = errors.New("StepFunc was not found")
)

// do runs the DoFunc on the nodes
func (s *Step) do(c TxnCtx) error {
	return runStepFuncOnNodes(s.DoFunc, c, s.Nodes)
}

// undo runs the UndoFunc on the nodes
func (s *Step) undo(c TxnCtx) error {
	if s.UndoFunc != "" {
		return runStepFuncOnNodes(s.UndoFunc, c, s.Nodes)
	}
	return nil
}

func runStepFuncOnNodes(name string, c TxnCtx, nodes []uuid.UUID) error {
	var (
		i    int
		node uuid.UUID
	)
	done := make(chan error)
	defer close(done)

	for i, node = range nodes {
		go runStepFuncOnNode(name, c, node, done)
	}

	// TODO: Need to properly aggregate results
	var err error
	for i >= 0 {
		err = <-done
		i--
	}
	return err
}

func runStepFuncOnNode(name string, c TxnCtx, node uuid.UUID, done chan<- error) {
	if uuid.Equal(node, gdctx.MyUUID) {
		done <- runStepFuncLocal(name, c)
	} else {
		done <- runStepFuncRemote(name, c, node)
	}
}

func runStepFuncLocal(name string, c TxnCtx) error {
	c.Logger().WithField("stepfunc", name).Debug("running step function")

	stepFunc, ok := GetStepFunc(name)
	if !ok {
		return ErrStepFuncNotFound
	}
	return stepFunc(c)
	//TODO: Results need to be aggregated
}

func runStepFuncRemote(step string, c TxnCtx, node uuid.UUID) error {
	rsp, err := RunStepOn(step, node, c)
	//TODO: Results need to be aggregated
	_ = rsp
	return err
}
