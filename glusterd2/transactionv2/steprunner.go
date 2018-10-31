package transaction

import (
	"context"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/glusterd2/transaction"

	"github.com/pborman/uuid"
)

// StepManager is an interface for running a step and also rollback step on local node
type StepManager interface {
	RunStep(ctx context.Context, step *transaction.Step, txnCtx transaction.TxnCtx) error
	RollBackStep(ctx context.Context, step *transaction.Step, txnCtx transaction.TxnCtx) error
	SyncStep(ctx context.Context, stepIndex int, txn *Txn) error
}

type stepManager struct {
	selfNodeID uuid.UUID
}

func newStepManager() StepManager {
	return &stepManager{
		selfNodeID: gdctx.MyUUID,
	}
}

func (sm *stepManager) shouldRunStep(step *transaction.Step) bool {
	if step.Skip {
		return false
	}

	for _, id := range step.Nodes {
		if uuid.Equal(sm.selfNodeID, id) {
			return true
		}
	}
	return false
}

// runStep synchronises the locally cached keys and values from the store
// before running the step function on node
func (sm *stepManager) runStep(ctx context.Context, stepName string, txnCtx transaction.TxnCtx) error {
	txnCtx.SyncCache()
	return transaction.RunStepFuncLocally(ctx, stepName, txnCtx)
}

// isPrevStepsExecutedOnNode reports that all pervious steps
// have been completed successfully on a given node
func (sm *stepManager) isPrevStepsExecutedOnNode(ctx context.Context, syncStepIndex int, nodeID uuid.UUID, txnID uuid.UUID, success chan<- struct{}) {
	txnManager := NewTxnManager(store.Store.Watcher)
	lastStepWatchChan := txnManager.WatchLastExecutedStep(ctx.Done(), txnID, nodeID)
	for {
		select {
		case <-ctx.Done():
			return
		case lastStep := <-lastStepWatchChan:
			if lastStep == syncStepIndex-1 {
				success <- struct{}{}
				return
			}
		}
	}
}

// SyncStep synchronises a step of given txn across all nodes of cluster
func (sm *stepManager) SyncStep(ctx context.Context, syncStepIndex int, txn *Txn) error {
	var (
		success         = make(chan struct{})
		syncCtx, cancel = context.WithTimeout(ctx, txnSyncTimeout)
	)
	defer cancel()

	for _, nodeID := range txn.Nodes {
		go sm.isPrevStepsExecutedOnNode(syncCtx, syncStepIndex, nodeID, txn.ID, success)
	}

	for range txn.Nodes {
		select {
		case <-syncCtx.Done():
			return errTxnSyncTimeout
		case <-success:
		}
	}
	return nil
}

// RollBackStep will rollback a given step on local node
func (sm *stepManager) RollBackStep(ctx context.Context, step *transaction.Step, txnCtx transaction.TxnCtx) error {
	if !sm.shouldRunStep(step) {
		return nil
	}

	if step.UndoFunc != "" {
		return sm.runStep(ctx, step.UndoFunc, txnCtx)
	}
	return nil
}

// RunStepRunStep will execute the step on local node
func (sm *stepManager) RunStep(ctx context.Context, step *transaction.Step, txnCtx transaction.TxnCtx) error {
	if !sm.shouldRunStep(step) {
		return nil
	}
	return sm.runStep(ctx, step.DoFunc, txnCtx)
}
