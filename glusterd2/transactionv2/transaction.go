// Package transaction implements a distributed transaction handling framework
package transaction

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/glusterd2/transaction"

	"github.com/coreos/etcd/clientv3"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/trace"
)

const (
	txnPrefix  = "transaction/"
	txnTimeOut = time.Minute * 3
)

// Txn is a set of steps
type Txn struct {
	locks transaction.Locks

	// Nodes is the union of the all the TxnStep.Nodes and is implicitly
	// set in Txn.Do(). This list is used to determine liveness of the
	// nodes before running the transaction steps.
	Nodes           []uuid.UUID         `json:"nodes"`
	StorePrefix     string              `json:"store_prefix"`
	ID              uuid.UUID           `json:"id"`
	ReqID           uuid.UUID           `json:"req_id"`
	Ctx             transaction.TxnCtx  `json:"ctx"`
	Steps           []*transaction.Step `json:"steps"`
	DontCheckAlive  bool                `json:"dont_check_alive"`
	DisableRollback bool                `json:"disable_rollback"`
	StartTime       time.Time           `json:"start_time"`
	TxnSpanCtx      trace.SpanContext   `json:"txn_span_ctx"`

	success   chan struct{}
	error     chan error
	succeeded bool
}

// NewTxn returns an initialized Txn without any steps
func NewTxn(ctx context.Context) *Txn {
	t := new(Txn)

	t.ID = uuid.NewRandom()
	t.ReqID = gdctx.GetReqID(ctx)
	t.locks = transaction.Locks{}
	t.StorePrefix = txnPrefix + t.ID.String() + "/"
	config := &transaction.TxnCtxConfig{
		LogFields: log.Fields{
			"txnid": t.ID.String(),
			"reqid": t.ReqID.String(),
		},
		StorePrefix: t.StorePrefix,
	}
	t.Ctx = transaction.NewCtx(config)
	spanCtx := trace.FromContext(ctx)
	t.TxnSpanCtx = spanCtx.SpanContext()
	t.Ctx.Logger().Debug("new transaction created")
	return t
}

// NewTxnWithLocks returns an empty Txn with locks obtained on given lockIDs
func NewTxnWithLocks(ctx context.Context, lockIDs ...string) (*Txn, error) {
	t := NewTxn(ctx)
	t.locks = transaction.Locks{}
	err := t.acquireClusterLocks(lockIDs...)
	return t, err
}

func (t *Txn) acquireClusterLocks(lockIDs ...string) error {
	if t.locks == nil {
		t.locks = transaction.Locks{}
	}

	for _, id := range lockIDs {
		logger := t.Ctx.Logger().WithField("lockID", id)
		logger.Debug("txn attempts to acquire cluster lock")
		if err := t.locks.Lock(id); err != nil {
			logger.WithError(err).Error("failed to obtain lock")
			t.releaseLocks()
			return err
		}
		logger.Debug("cluster lock acquired")
	}

	return nil
}

func (t *Txn) releaseLocks() {
	t.locks.UnLock(context.Background())
}

// Done releases any obtained locks and cleans up the transaction namespace
// Done must be called after a transaction ends
func (t *Txn) Done() {
	defer t.releaseLocks()
	if !t.succeeded {
		return
	}

	t.Ctx.Logger().Info("transaction succeeded on all nodes")
	t.removeContextData()

	if err := GlobalTxnManager.RemoveTransaction(t.ID); err != nil {
		t.Ctx.Logger().WithError(err).Error("failed to remove txn data from pending-transaction namespace")
	}
}

func (t *Txn) removeContextData() {
	if _, err := store.Delete(context.TODO(), t.StorePrefix, clientv3.WithPrefix()); err != nil {
		t.Ctx.Logger().WithError(err).WithField("key",
			t.StorePrefix).Error("Failed to remove transaction namespace from store")
	}

}

func (t *Txn) checkAlive() error {
	for _, node := range t.Nodes {
		// TODO: Using prefixed query, get all alive nodes in a single etcd query
		if _, online := store.Store.IsNodeAlive(node); !online {
			return fmt.Errorf("node %s is probably down", node.String())
		}
	}
	return nil
}

// Do runs the transaction on the cluster
func (t *Txn) Do() error {
	var (
		stop  = make(chan struct{})
		timer = time.NewTimer(txnTimeOut)
	)

	{
		t.success = make(chan struct{})
		t.error = make(chan error)
		t.StartTime = time.Now()
	}

	defer timer.Stop()

	if len(t.Nodes) == 0 {
		for _, s := range t.Steps {
			t.Nodes = append(t.Nodes, s.Nodes...)
		}
	}
	t.Nodes = nodesUnion(t.Nodes)

	if !t.DontCheckAlive {
		if err := t.checkAlive(); err != nil {
			return err
		}
	}

	t.Ctx.Logger().Debug("Starting transaction")

	go t.waitForCompletion(stop)
	defer close(stop)

	GlobalTxnManager.UpDateTxnStatus(TxnStatus{State: txnPending, TxnID: t.ID}, t.ID, t.Nodes...)
	GlobalTxnManager.UpdateLastExecutedStep(-1, t.ID, t.Nodes...)

	// commit txn.Ctx.Set()s done in REST handlers to the store
	if err := t.Ctx.Commit(); err != nil {
		return err
	}

	t.Ctx.Logger().Debug("adding txn to store")
	if err := GlobalTxnManager.AddTxn(t); err != nil {
		return err
	}
	t.Ctx.Logger().Debug("waiting for completion of transaction")

	failureAction := func(err error) {
		t.Ctx.Logger().WithError(err).Error("error in executing txn, marking as failure")
		txnStatus := TxnStatus{State: txnFailed, TxnID: t.ID, Reason: err.Error()}
		GlobalTxnManager.UpDateTxnStatus(txnStatus, t.ID, t.Nodes...)
	}

	select {
	case <-t.success:
		t.succeeded = true
	case err := <-t.error:
		failureAction(err)
		return err
	case <-timer.C:
		failureAction(errTxnTimeout)
		return errTxnTimeout
	}

	return nil
}

// notifyState will send a notification on `success` chan if txn got marked as succeeded on given
// nodeID. In case txn got failed on the given nodeID then it will send a notification on Txn.error
// chan.
func (t *Txn) notifyState(nodeID uuid.UUID, success chan<- struct{}, stopCh <-chan struct{}) {
	txnStatusChan := GlobalTxnManager.WatchTxnStatus(stopCh, t.ID, nodeID)

	for {
		select {
		case <-stopCh:
			return
		case status := <-txnStatusChan:
			log.WithFields(log.Fields{
				"nodeId": nodeID.String(),
				"status": fmt.Sprintf("%+v", status),
			}).Debug("state received")

			if status.State == txnSucceeded {
				success <- struct{}{}
				return
			} else if status.State == txnFailed {
				t.error <- errors.New(status.Reason)
				return
			}
		}
	}
}

// waitForCompletion will wait for transaction to complete on all nodes.
// If txn got marked as succeeded on all nodes then it will send a notification
// on Txn.success chan.
func (t *Txn) waitForCompletion(stopCh <-chan struct{}) {
	var successChan = make(chan struct{})

	for _, nodeID := range t.Nodes {
		go t.notifyState(nodeID, successChan, stopCh)
	}

	for range t.Nodes {
		select {
		case <-stopCh:
			return
		case <-successChan:
		}
	}
	t.success <- struct{}{}
}

// nodesUnion removes duplicate nodes
func nodesUnion(nodes []uuid.UUID) []uuid.UUID {
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			if uuid.Equal(nodes[i], nodes[j]) {
				nodes = append(nodes[:j], nodes[j+1:]...)
				j--
			}
		}
	}
	return nodes
}

// FilterNonFailedTxn will return txns which are not marked as failed
func FilterNonFailedTxn(txns []*Txn) []*Txn {
	var nonFailedTxns []*Txn
	for _, txn := range txns {
		txnStatus, err := GlobalTxnManager.GetTxnStatus(txn.ID, gdctx.MyUUID)
		if err == nil && txnStatus.State.Valid() && txnStatus.State != txnFailed {
			nonFailedTxns = append(nonFailedTxns, txn)
		}
	}
	return nonFailedTxns
}
