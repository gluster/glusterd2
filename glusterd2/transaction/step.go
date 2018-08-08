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
func (s *Step) do(origCtx context.Context, ctx TxnCtx) error {
	return runStepFuncOnNodes(origCtx, s.DoFunc, ctx, s.Nodes)
}

// undo runs the UndoFunc on the nodes
func (s *Step) undo(ctx TxnCtx) error {
	if s.UndoFunc != "" {
		return runStepFuncOnNodes(nil, s.UndoFunc, ctx, s.Nodes)
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

func runStepFuncOnNodes(origCtx context.Context, stepName string, ctx TxnCtx, nodes []uuid.UUID) error {

	respCh := make(chan stepPeerResp, len(nodes))
	defer close(respCh)

	for _, node := range nodes {
		go runStepFuncOnNode(origCtx, stepName, ctx, node, respCh)
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
			ctx.Logger().WithError(peerResp.Error).WithFields(log.Fields{
				"step": stepName, "node": peerResp.PeerID,
			}).Error("Step failed on node.")
		}
		resp.Resps = append(resp.Resps, peerResp)
	}

	if resp.errCount != 0 {
		return resp
	}

	return nil
}

func runStepFuncOnNode(origCtx context.Context, stepName string, ctx TxnCtx, node uuid.UUID, respCh chan<- stepPeerResp) {

	ctx.Logger().WithFields(log.Fields{
		"step": stepName, "node": node,
	}).Debug("Running step on node.")

	var err error
	if uuid.Equal(node, gdctx.MyUUID) {
		err = runStepFuncLocally(origCtx, stepName, ctx)
	} else {
		// remote node
		err = runStepOn(origCtx, stepName, node, ctx)
	}

	respCh <- stepPeerResp{node, err}
}

func runStepFuncLocally(origCtx context.Context, stepName string, ctx TxnCtx) error {

	var err error

	if origCtx != nil {
		reqID := ctx.GetTxnReqID()
		spanName := stepName + " ReqID:" + reqID
		_, span := trace.StartSpan(origCtx, spanName)
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
