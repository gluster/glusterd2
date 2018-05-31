package transaction

import (
	"errors"
	"fmt"

	"github.com/gluster/glusterd2/glusterd2/gdctx"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
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

type stepResp struct {
	node uuid.UUID
	step string
	err  error
}

func runStepFuncOnNodes(stepName string, c TxnCtx, nodes []uuid.UUID) error {

	respCh := make(chan stepResp, len(nodes))
	defer close(respCh)

	for _, node := range nodes {
		go runStepFuncOnNode(stepName, c, node, respCh)
	}

	// Ideally, we have to cancel the pending go-routines on first error
	// response received from any of the nodes. But that's really tricky
	// to do. Serializing sequentially is the easiest fix but we lose
	// concurrency. Instead, we let the do() function run on all nodes.

	var lastErrResp stepResp
	errCount := 0
	var resp stepResp
	for range nodes {
		resp = <-respCh
		if resp.err != nil {
			errCount++
			c.Logger().WithFields(log.Fields{
				"step": resp.step, "node": resp.node,
			}).WithError(resp.err).Error("Step failed on node.")
			lastErrResp = resp
		}
	}

	if errCount != 0 {
		return fmt.Errorf("Step %s failed on %d nodes. "+
			"Last error: Step func %s failed on %s with error: %s",
			stepName, errCount, lastErrResp.step, lastErrResp.node, lastErrResp.err)
	}

	return nil
}

func runStepFuncOnNode(stepName string, c TxnCtx, node uuid.UUID, respCh chan<- stepResp) {

	c.Logger().WithFields(log.Fields{
		"step": stepName, "node": node,
	}).Debug("Running step on node.")

	var err error
	if uuid.Equal(node, gdctx.MyUUID) {
		// this (local) node
		if stepFunc, ok := getStepFunc(stepName); ok {
			err = stepFunc(c)
		} else {
			err = ErrStepFuncNotFound
		}
	} else {
		// remote node
		err = runStepOn(stepName, node, c)
	}

	respCh <- stepResp{node, stepName, err}
}
