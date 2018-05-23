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
	Skip     bool
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

	// Ideally, we have to cancel the pending go-routines on first error
	// response received from any of the nodes. But that's really tricky
	// to do. Serializing sequentially is the easiest fix but we lose
	// concurrency. Instead, we let the do() function run on all nodes.

	var err, lastErr error
	for i >= 0 {
		err = <-done
		if err != nil {
			// TODO: Need to properly aggregate results and
			// check which node returned which error response.
			lastErr = err
		}
		i--
	}

	return lastErr
}

func runStepFuncOnNode(name string, c TxnCtx, node uuid.UUID, done chan<- error) {
	if uuid.Equal(node, gdctx.MyUUID) {
		done <- runStepFuncLocal(name, c)
	} else {
		done <- runStepOn(name, node, c)
	}
}

func runStepFuncLocal(name string, c TxnCtx) error {
	c.Logger().WithField("stepfunc", name).Debug("running step function")

	stepFunc, ok := getStepFunc(name)
	if !ok {
		return ErrStepFuncNotFound
	}
	return stepFunc(c)
	//TODO: Results need to be aggregated
}
