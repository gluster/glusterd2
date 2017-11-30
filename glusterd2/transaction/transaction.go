// Package transaction implements a distributed transaction handling framework
package transaction

import (
	"context"
	"expvar"
	"fmt"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/store"

	"github.com/coreos/etcd/clientv3"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	txnPrefix = "transaction/"
)

var expTxn = expvar.NewMap("txn")

// Txn is a set of steps
type Txn struct {
	id    uuid.UUID
	reqID uuid.UUID
	Ctx   TxnCtx
	Steps []*Step
	// Nodes should be a union of the all the TxnStep.Nodes and must be set
	// before calling Txn.Do(). This is currently only used to determine
	// liveness of the nodes before running the transaction steps.
	Nodes []uuid.UUID
}

// NewTxn returns an initialized Txn without any steps
func NewTxn(ctx context.Context) *Txn {
	t := new(Txn)
	t.id = uuid.NewRandom()
	t.reqID = gdctx.GetReqID(ctx)
	prefix := txnPrefix + t.id.String()
	t.Ctx = NewCtxWithLogFields(log.Fields{
		"txnid": t.id.String(),
		"reqid": t.reqID.String(),
	}).WithPrefix(prefix)

	return t
}

// Cleanup cleans the leftovers after a transaction ends
func (t *Txn) Cleanup() {
	store.Store.Delete(context.TODO(), t.Ctx.Prefix(), clientv3.WithPrefix())
	expTxn.Add("initiated_txn_in_progress", -1)
}

// Do runs the transaction on the cluster
func (t *Txn) Do() (TxnCtx, error) {
	t.Ctx.Logger().Debug("Starting transaction")

	// verify that all nodes are online
	for _, node := range t.Nodes {
		if !store.Store.IsNodeAlive(node) {
			return nil, fmt.Errorf("node %s is probably down", node.String())
		}
	}

	expTxn.Add("initiated_txn_in_progress", 1)

	//Do the steps
	for i, s := range t.Steps {
		if e := s.do(t.Ctx); e != nil {
			t.Ctx.Logger().WithError(e).Error("Transaction failed, rolling back changes")
			t.undo(i)
			return nil, e
		}
	}

	return t.Ctx, nil
}

// undo undoes a transaction and will be automatically called by Perform if any step fails.
// The Steps are undone in the reverse order, from the failed step.
func (t *Txn) undo(n int) {
	for i := n; i >= 0; i-- {
		t.Steps[i].undo(t.Ctx)
	}
}
