// Package transaction implements a distributed transaction handling framework
package transaction

import (
	"context"
	"expvar"
	"fmt"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/store"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	txnPrefix = "transaction/"
)

var expTxn = expvar.NewMap("txn")

// Txn is a set of steps
type Txn struct {
	id          uuid.UUID
	locks       map[string]*concurrency.Mutex
	reqID       uuid.UUID
	storePrefix string

	Ctx             TxnCtx
	Steps           []*Step
	DontCheckAlive  bool
	DisableRollback bool
	// Nodes is the union of the all the TxnStep.Nodes and is implicitly
	// set in Txn.Do(). This list is used to determine liveness of the
	// nodes before running the transaction steps.
	Nodes []uuid.UUID
}

// NewTxn returns an initialized Txn without any steps
func NewTxn(ctx context.Context) *Txn {
	t := new(Txn)

	t.id = uuid.NewRandom()
	t.reqID = gdctx.GetReqID(ctx)
	t.locks = make(map[string]*concurrency.Mutex)
	t.storePrefix = txnPrefix + t.id.String() + "/"
	config := &txnCtxConfig{
		LogFields: log.Fields{
			"txnid": t.id.String(),
			"reqid": t.reqID.String(),
		},
		StorePrefix: t.storePrefix,
	}
	t.Ctx = newCtx(config)

	t.Ctx.Logger().Debug("new transaction created")
	return t
}

// NewTxnWithLocks returns an empty Txn with locks obtained on given lockIDs
func NewTxnWithLocks(ctx context.Context, lockIDs ...string) (*Txn, error) {
	t := NewTxn(ctx)

	for _, id := range lockIDs {
		if err := t.Lock(id); err != nil {
			t.Done()
			return nil, err
		}
	}

	return t, nil
}

// Cleanup cleans the leftovers after a transaction ends
// TODO: Remove this function
func (t *Txn) Cleanup() {
	if _, err := store.Store.Delete(context.TODO(), t.storePrefix, clientv3.WithPrefix()); err != nil {
		t.Ctx.Logger().WithError(err).WithField("key",
			t.storePrefix).Error("Failed to remove transaction namespace from store")
	}
	expTxn.Add("initiated_txn_in_progress", -1)
}

// Done releases any obtained locks and cleans up the transaction namespace
// Done must be called after a transaction ends
func (t *Txn) Done() {
	// Release obtained locks
	for _, locker := range t.locks {
		locker.Unlock(context.Background())
	}
	t.Cleanup()
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
		if !store.Store.IsNodeAlive(node) {
			return fmt.Errorf("node %s is probably down", node.String())
		}
	}

	return nil
}

// Do runs the transaction on the cluster
func (t *Txn) Do() error {
	if !t.DontCheckAlive {
		if err := t.checkAlive(); err != nil {
			return err
		}
	}

	t.Ctx.Logger().Debug("Starting transaction")
	expTxn.Add("initiated_txn_in_progress", 1)

	// commit txn.Ctx.Set()s done in REST handlers to the store
	if err := t.Ctx.commit(); err != nil {
		return err
	}

	for i, s := range t.Steps {
		if s.Skip {
			continue
		}

		if err := s.do(t.Ctx); err != nil {
			if !t.DisableRollback {
				t.Ctx.Logger().WithError(err).Error("Transaction failed, rolling back changes")
				t.undo(i)
			}
			return err
		}
	}

	return nil
}

// undo undoes a transaction and will be automatically called by Perform if any step fails.
// The Steps are undone in the reverse order, from the failed step.
func (t *Txn) undo(n int) {
	for i := n; i >= 0; i-- {
		if t.Steps[i].Skip {
			continue
		}
		t.Steps[i].undo(t.Ctx)
	}
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
