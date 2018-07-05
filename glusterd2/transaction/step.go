package transaction

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/pkg/api"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/trace"
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
	ErrStepFuncNotFound = errors.New("stepFunc was not found")
)

// do runs the DoFunc on the nodes
func (s *Step) do(ctx TxnCtx, origCtx ...context.Context) error {
	return runStepFuncOnNodes(s.DoFunc, ctx, s.Nodes, origCtx...)
}

// undo runs the UndoFunc on the nodes
func (s *Step) undo(ctx TxnCtx) error {
	if s.UndoFunc != "" {
		return runStepFuncOnNodes(s.UndoFunc, ctx, s.Nodes)
	}
	return nil
}

// stepPeerResp is response from a single peer that runs a step
type stepPeerResp struct {
	PeerID uuid.UUID
	Error  error
}

// stepResp contains response from multiple peers that run a step and the type
// implements the `api.ErrorResponse` interface
type stepResp struct {
	Step     string
	Resps    []stepPeerResp
	errCount int
}

func (r stepResp) Error() string {
	return fmt.Sprintf("Step %s failed on %d nodes", r.Step, r.errCount)
}

func (r stepResp) Response() api.ErrorResp {

	var apiResp api.ErrorResp
	for _, resp := range r.Resps {
		if resp.Error == nil {
			continue
		}

		apiResp.Errors = append(apiResp.Errors, api.HTTPError{
			Code:    int(api.ErrTxnStepFailed),
			Message: api.ErrorCodeMap[api.ErrTxnStepFailed],
			Fields: map[string]string{
				"peer-id": resp.PeerID.String(),
				"step":    r.Step,
				"error":   resp.Error.Error()},
		})
	}

	return apiResp
}

func (r stepResp) Status() int {
	return http.StatusInternalServerError
}

func runStepFuncOnNodes(stepName string, ctx TxnCtx, nodes []uuid.UUID, origCtx ...context.Context) error {

	respCh := make(chan stepPeerResp, len(nodes))
	defer close(respCh)

	for _, node := range nodes {
		go runStepFuncOnNode(stepName, ctx, node, respCh, origCtx...)
	}

	// Ideally, we have to cancel the pending go-routines on first error
	// response received from any of the nodes. But that's really tricky
	// to do. Serializing sequentially is the easiest fix but we lose
	// concurrency. Instead, we let the do() function run on all nodes.

	resp := stepResp{
		Step:  stepName,
		Resps: make([]stepPeerResp, len(nodes)),
	}

	var peerResp stepPeerResp
	for range nodes {
		peerResp = <-respCh
		if peerResp.Error != nil {
			resp.errCount++
			ctx.Logger().WithFields(log.Fields{
				"step": stepName, "node": peerResp.PeerID,
			}).WithError(peerResp.Error).Error("Step failed on node.")
		}
		resp.Resps = append(resp.Resps, peerResp)
	}

	if resp.errCount != 0 {
		return resp
	}

	return nil
}

func runStepFuncOnNode(stepName string, ctx TxnCtx, node uuid.UUID, respCh chan<- stepPeerResp, origCtx ...context.Context) {

	ctx.Logger().WithFields(log.Fields{
		"step": stepName, "node": node,
	}).Debug("Running step on node.")

	var err error
	if uuid.Equal(node, gdctx.MyUUID) {
		err = runStepFuncLocally(stepName, ctx, origCtx...)
	} else {
		// remote node
		err = runStepOn(stepName, node, ctx)
	}

	respCh <- stepPeerResp{node, err}
}

func runStepFuncLocally(stepName string, ctx TxnCtx, origCtx ...context.Context) error {

	var err error

	if origCtx != nil {
		txCtx := origCtx[0]
		reqID := ctx.GetTxnReqID()
		spanName := stepName + " ReqID:" + reqID
		txCtx, span := trace.StartSpan(txCtx, spanName)
		defer span.End()
	}

	stepFunc, ok := getStepFunc(stepName)
	if ok {
		if err = stepFunc(ctx); err == nil {
			// if step function executes successfully, commit the
			// results to the store
			err = ctx.commit()
		}
	} else {
		err = ErrStepFuncNotFound
	}

	return err
}
