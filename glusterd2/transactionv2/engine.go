package transaction

import (
	"context"
	"errors"
	"github.com/gluster/glusterd2/glusterd2/store"
	"sync"
	"time"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	// PendingTxnPrefix is the etcd namespace into which all pending txn will be stored
	PendingTxnPrefix = "pending-transaction/"
	txnSyncTimeout   = time.Second * 10
)

// transactionEngine is responsible for executing newly added txn
var transactionEngine *Engine

// Engine executes the given transaction across the cluster.
// It makes use of etcd as the means of communication between nodes.
type Engine struct {
	stop        chan struct{}
	stopOnce    sync.Once
	selfNodeID  uuid.UUID
	stepManager StepManager
	txnManager  TxnManager
}

// NewEngine creates a TxnEngine
func NewEngine() *Engine {
	return &Engine{
		stop:        make(chan struct{}),
		selfNodeID:  gdctx.MyUUID,
		stepManager: newStepManager(),
		txnManager:  NewTxnManager(store.Store.Watcher),
	}
}

// Run will start running the TxnEngine and wait for txn Engine to be stopped.
func (txnEng *Engine) Run() {
	log.Info("running txn engine")

	go UntilStop(txnEng.HandleTransaction, 0, txnEng.stop)
	go UntilStop(txnEng.HandleFailedTxn, 0, txnEng.stop)

	<-txnEng.stop
	log.Info("txn engine stopped")
}

// HandleTransaction executes newly added txn to the store. It will keep watching on
// `pending-transaction` namespace, if a new txn is added to the namespace then it will
// execute that txn.
func (txnEng *Engine) HandleTransaction() {
	txnChan := txnEng.txnManager.WatchTxn(txnEng.stop)

	for {
		select {
		case <-txnEng.stop:
			return
		case txn, ok := <-txnChan:
			if !ok {
				return
			}
			txn.Ctx.Logger().Debugf("received a pending txn")
			go txnEng.Execute(context.Background(), txn)
		}
	}
}

func (txnEng *Engine) isInitiator(txn *Txn) bool {
	return uuid.Equal(txn.Initiator, txnEng.selfNodeID)
}

// Execute will run a given txn
func (txnEng *Engine) Execute(ctx context.Context, txn *Txn) {
	var shouldExecute bool
	for _, node := range txn.Nodes {
		if uuid.Equal(node, txnEng.selfNodeID) {
			txn.Ctx.Logger().WithField("peerID", txnEng.selfNodeID.String()).Info("executing txn on peer")
			shouldExecute = true
			break
		}
	}

	if !shouldExecute {
		txn.Ctx.Logger().WithField("peerID", txnEng.selfNodeID.String()).Info("peer is not involved in this txn")
		return
	}

	if txnEng.isInitiator(txn) {
		if err := WithClusterLocks(txn.Locks...)(txn); err != nil {
			txn.Ctx.Logger().WithError(err).Error("error in acquiring cluster lock")
			return
		}
		defer txn.releaseLocks()
	}

	txnStatus, err := txnEng.txnManager.GetTxnStatus(txn.ID, txnEng.selfNodeID)
	if err != nil {
		txn.Ctx.Logger().WithError(err).Error("error in getting txn status")
		return
	}

	switch txnStatus.State {
	case txnPending:
		if err := txnEng.executePendingTxn(ctx, txn); err != nil {
			status := TxnStatus{State: txnFailed, Reason: err.Error(), TxnID: txn.ID}
			txnEng.txnManager.UpDateTxnStatus(status, txn.ID, txnEng.selfNodeID)
			return
		}
		txn.Ctx.Logger().Info("txn succeeded")
		txnEng.txnManager.UpDateTxnStatus(TxnStatus{State: txnSucceeded, TxnID: txn.ID}, txn.ID, txnEng.selfNodeID)
	}
	return
}

func (txnEng *Engine) executePendingTxn(ctx context.Context, txn *Txn) error {
	var (
		stopch        = make(chan struct{})
		txnStatusChan = txnEng.txnManager.WatchTxnStatus(stopch, txn.ID, txnEng.selfNodeID)
		updateOnce    = &sync.Once{}
	)
	defer close(stopch)

	txn.Ctx.Logger().Infof("transaction started on node")

	for i, step := range txn.Steps {
		if step.Sync {
			txn.Ctx.Logger().WithField("stepname", step.DoFunc).Debug("synchronizing txn step")
			if err := txnEng.stepManager.SyncStep(ctx, i, txn); err != nil {
				txn.Ctx.Logger().WithFields(log.Fields{
					"error":    err,
					"stepname": step.DoFunc,
				}).Error("encounter an error in synchronizing txn step")
				return err
			}
			txn.Ctx.Logger().Debug("transaction got synchronized")
		}

		if err := txnEng.stepManager.RunStep(ctx, step, txn.Ctx); err != nil {
			txn.Ctx.Logger().WithFields(log.Fields{
				"error":    err,
				"stepname": step.DoFunc,
			}).Error("failed in executing step ")
			txnEng.txnManager.UpdateLastExecutedStep(i, txn.ID, txnEng.selfNodeID)
			return err
		}

		select {
		case <-ctx.Done():
			txnEng.txnManager.UpdateLastExecutedStep(i, txn.ID, txnEng.selfNodeID)
			return ctx.Err()
		case status := <-txnStatusChan:
			if status.State == txnFailed {
				txnEng.txnManager.UpdateLastExecutedStep(i, txn.ID, txnEng.selfNodeID)
				return errors.New(status.Reason)
			}
		default:
		}

		txnEng.txnManager.UpdateLastExecutedStep(i, txn.ID, txnEng.selfNodeID)
		updateOnce.Do(func() {
			txnEng.txnManager.UpDateTxnStatus(TxnStatus{State: txnRunning, TxnID: txn.ID}, txn.ID, txnEng.selfNodeID)
		})
	}
	return nil
}

// HandleFailedTxn keep on watching store for failed txn. If it receives any failed
// txn then it will rollback all executed steps of that txn.
func (txnEng *Engine) HandleFailedTxn() {
	failedTxnChan := txnEng.txnManager.WatchFailedTxn(txnEng.stop, txnEng.selfNodeID)

	for {
		select {
		case <-txnEng.stop:
			return
		case failedTxn, ok := <-failedTxnChan:
			if !ok {
				return
			}

			lastStepIndex, err := txnEng.txnManager.GetLastExecutedStep(failedTxn.ID, txnEng.selfNodeID)
			if err != nil || lastStepIndex == -1 {
				continue
			}
			failedTxn.Ctx.Logger().Debugf("received a failed txn, rolling back changes")

			for i := lastStepIndex; i >= 0; i-- {
				err := txnEng.stepManager.RollBackStep(context.Background(), failedTxn.Steps[i], failedTxn.Ctx)
				if err != nil {
					failedTxn.Ctx.Logger().WithError(err).WithField("step", failedTxn.Steps[i]).Error("failed in rolling back step")
				}
			}
			txnEng.txnManager.UpdateLastExecutedStep(-1, failedTxn.ID, txnEng.selfNodeID)
		}
	}
}

// Stop will stop a running Txn Engine
func (txnEng *Engine) Stop() {
	log.Info("stopping txn engine")
	txnEng.stopOnce.Do(func() {
		close(txnEng.stop)
	})
}

// StartTxnEngine creates a new Txn Engine and starts running it
func StartTxnEngine() {
	transactionEngine = NewEngine()
	GlobalTxnManager = NewTxnManager(store.Store.Watcher)
	go transactionEngine.Run()
}

// StopTxnEngine stops the Txn Engine
func StopTxnEngine() {
	if transactionEngine != nil {
		transactionEngine.Stop()
	}
}
