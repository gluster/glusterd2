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
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	txnPrefix  = "transaction/"
	txnTimeOut = time.Second * 15
)

// TxnOptFunc receives a Txn and overrides its members
type TxnOptFunc func(*Txn) error

// Txn is a set of steps
type Txn struct {
	locks transaction.Locks

	// Nodes is the union of the all the TxnStep.Nodes and is implicitly
	// set in Txn.Do(). This list is used to determine liveness of the
	// nodes before running the transaction steps.
	Nodes           []uuid.UUID         `json:"nodes"`
	StorePrefix     string              `json:"store_prefix"`
	ID              uuid.UUID           `json:"id"`
	Locks           []string            `json:"locks"`
	ReqID           uuid.UUID           `json:"req_id"`
	Ctx             transaction.TxnCtx  `json:"ctx"`
	Steps           []*transaction.Step `json:"steps"`
	DontCheckAlive  bool                `json:"dont_check_alive"`
	DisableRollback bool                `json:"disable_rollback"`
	Initiator       uuid.UUID           `json:"initiator"`
	StartTime       time.Time           `json:"start_time"`

	success   chan struct{}
	error     chan error
	succeeded bool
}

// NewTxn returns an initialized Txn without any steps
func NewTxn(ctx context.Context) *Txn {
	t := new(Txn)

	t.ID = uuid.NewRandom()
	t.ReqID = gdctx.GetReqID(ctx)
	t.locks = make(map[string]*concurrency.Mutex)
	t.StorePrefix = txnPrefix + t.ID.String() + "/"
	config := &transaction.TxnCtxConfig{
		LogFields: log.Fields{
			"txnid": t.ID.String(),
			"reqid": t.ReqID.String(),
		},
		StorePrefix: t.StorePrefix,
	}
	t.Ctx = transaction.NewCtx(config)
	t.Initiator = gdctx.MyUUID
	t.Ctx.Logger().Debug("new transaction created")
	return t
}

// NewTxnWithLocks returns an empty Txn with locks obtained on given lockIDs
func NewTxnWithLocks(ctx context.Context, lockIDs ...string) (*Txn, error) {
	t := NewTxn(ctx)
	t.Locks = lockIDs
	return t, nil
}

// WithClusterLocks obtains a cluster wide locks on given IDs for a txn
func WithClusterLocks(lockIDs ...string) TxnOptFunc {
	return func(t *Txn) error {
		for _, id := range lockIDs {
			logger := t.Ctx.Logger().WithField("lockID", id)
			logger.Debug("attempting to obtain lock")
			if err := t.locks.Lock(id); err != nil {
				logger.WithError(err).Error("failed to obtain lock")
				t.releaseLocks()
				return err
			}
			logger.Debug("lock obtained")
		}
		return nil
	}
}

func (t *Txn) releaseLocks() {
	t.locks.UnLock(context.Background())
}

// Done releases any obtained locks and cleans up the transaction namespace
// Done must be called after a transaction ends
func (t *Txn) Done() {
	if t.succeeded {
		t.done()
		t.releaseLocks()
		GlobalTxnManager.RemoveTransaction(t.ID)
		t.Ctx.Logger().Info("txn succeeded on all nodes, txn data cleaned up from store")
	}
}

func (t *Txn) done() {
	if _, err := store.Delete(context.TODO(), t.StorePrefix, clientv3.WithPrefix()); err != nil {
		t.Ctx.Logger().WithError(err).WithField("key",
			t.StorePrefix).Error("Failed to remove transaction namespace from store")
	}

}

func (t *Txn) checkAlive() error {

	if len(t.Nodes) == 0 {
		for _, s := range t.Steps {
			t.Nodes = append(t.Nodes, s.Nodes...)
		}
	}
	t.Nodes = nodesUnion(t.Nodes)

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

	if !t.DontCheckAlive {
		if err := t.checkAlive(); err != nil {
			return err
		}
	}

	t.Ctx.Logger().Debug("Starting transaction")

	go t.waitForCompletion(stop)

	GlobalTxnManager.UpDateTxnStatus(TxnStatus{State: txnPending, TxnID: t.ID}, t.ID, t.Nodes...)

	// commit txn.Ctx.Set()s done in REST handlers to the store
	if err := t.Ctx.Commit(); err != nil {
		return err
	}

	t.Ctx.Logger().Debug("adding txn to store")
	if err := GlobalTxnManager.AddTxn(t); err != nil {
		return err
	}
	t.Ctx.Logger().Debug("waiting for txn to be cleaned up")

	select {
	case <-t.success:
		close(stop)
		t.succeeded = true
	case err := <-t.error:
		t.Ctx.Logger().WithError(err).Error("error in executing txn, marking as failure")
		close(stop)
		txnStatus := TxnStatus{State: txnFailed, TxnID: t.ID, Reason: err.Error()}
		GlobalTxnManager.UpDateTxnStatus(txnStatus, t.ID, t.Nodes...)
		return err
	case <-timer.C:
		t.Ctx.Logger().Error("time out in cleaning txn, marking as failure")
		close(stop)
		for _, nodeID := range t.Nodes {
			txnStatus := TxnStatus{State: txnFailed, TxnID: t.ID, Reason: "txn timed out"}
			GlobalTxnManager.UpDateTxnStatus(txnStatus, t.ID, nodeID)
		}
		return errTxnTimeout
	}

	return nil
}

func (t *Txn) isNodeSucceded(nodeID uuid.UUID, success chan<- struct{}, stopCh <-chan struct{}) {
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

func (t *Txn) waitForCompletion(stopCh <-chan struct{}) {
	var successChan = make(chan struct{})

	for _, nodeID := range t.Nodes {
		go t.isNodeSucceded(nodeID, successChan, stopCh)
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
