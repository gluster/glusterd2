package transaction

import (
	"context"
	"sync"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/store"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	// PendingTxnPrefix is the etcd namespace into which all pending txn will be stored
	PendingTxnPrefix = "pending-transaction/"
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
	executor    Executor
}

// NewEngine creates a TxnEngine
func NewEngine() *Engine {
	return &Engine{
		stop:        make(chan struct{}),
		selfNodeID:  gdctx.MyUUID,
		stepManager: newStepManager(),
		txnManager:  NewTxnManager(store.Store.Watcher),
		executor:    NewExecutor(),
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
	pendingTxnChan := make(chan *Txn)

	// fetching list of all non-failed transaction which are
	// already present in store. If a transaction is not marked
	// as fail then we need to resume all those transaction.
	go func(txnChan chan<- *Txn) {
		txns := FilterNonFailedTxn(txnEng.txnManager.GetTxns())
		for _, txn := range txns {
			txnChan <- txn
		}
	}(pendingTxnChan)

	// watching store for newly added txns
	txnChan := txnEng.txnManager.WatchTxn(txnEng.stop)

	for {
		select {
		case <-txnEng.stop:
			return
		case txn := <-pendingTxnChan:
			go txnEng.Execute(context.Background(), txn)
		case txn, ok := <-txnChan:
			if !ok {
				return
			}
			go txnEng.Execute(context.Background(), txn)
		}
	}
}

// Execute will run a given txn
func (txnEng *Engine) Execute(ctx context.Context, txn *Txn) {
	status, err := txnEng.txnManager.GetTxnStatus(txn.ID, txnEng.selfNodeID)
	if err != nil {
		return
	}

	txn.Ctx.Logger().WithField("state", status.State).Debug("received a transaction")

	switch status.State {
	case txnPending:
		if err := txnEng.executor.Execute(ctx, txn); err != nil {
			txn.Ctx.Logger().WithError(err).Error("error in executing transaction")
		}
	case txnRunning:
		if err := txnEng.executor.Resume(ctx, txn); err != nil {
			txn.Ctx.Logger().WithError(err).Error("error in resuming transaction")
		}
	}

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
			go txnEng.startRollBack(failedTxn)
		}
	}
}

func (txnEng *Engine) startRollBack(failedTxn *Txn) {
	lastStepIndex, err := txnEng.txnManager.GetLastExecutedStep(failedTxn.ID, txnEng.selfNodeID)
	if err != nil || lastStepIndex == -1 {
		return
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
