// Package transaction implements a distributed transaction handling framework
package transaction

import (
	"context"

	"github.com/gluster/glusterd2/gdctx"
	"github.com/gluster/glusterd2/store"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/clientv3"
	"github.com/pborman/uuid"
)

const (
	txnPrefix = store.GlusterPrefix + "transaction/"
)

// Txn is a set of steps
//
// Nodes is a union of the all the TxnStep.Nodes
type Txn struct {
	// TODO: Any good reason for this to be not just string ?
	ID    uuid.UUID
	Ctx   TxnCtx
	Steps []*Step
	Nodes []uuid.UUID
}

// NewTxn returns an initialized Txn without any steps
func NewTxn(id string) *Txn {
	t := new(Txn)

	if uuid.Parse(id) != nil {
		t.ID = uuid.Parse(id)
	} else {
		t.ID = uuid.NewRandom()
	}

	prefix := txnPrefix + t.ID.String()
	t.Ctx = NewCtxWithLogFields(log.Fields{
		"reqid": t.ID.String(),
	}).WithPrefix(prefix)

	return t
}

// NewTxnWithLoggingContext creates a Txn with a Context with given logging fields
func NewTxnWithLoggingContext(f log.Fields, id string) *Txn {
	t := NewTxn(id)
	prefix := txnPrefix + t.ID.String()
	t.Ctx = NewCtxWithLogFields(log.Fields{
		"reqid": t.ID.String(),
	}).WithPrefix(prefix).WithLogFields(f)

	return t
}

// Cleanup cleans the leftovers after a transaction ends
func (t *Txn) Cleanup() {
	gdctx.Store.Delete(context.TODO(), t.Ctx.Prefix(), clientv3.WithPrefix())
}

// Do runs the transaction on the cluster
func (t *Txn) Do() (TxnCtx, error) {
	t.Ctx.Logger().Debug("Starting transaction")

	//First verify all nodes are online
	for range t.Nodes {
		/*
			if !Online(n) {
				return error
			}
		*/
	}

	//Do the steps
	for i, s := range t.Steps {
		//TODO: Renable (correctly) if All/Leader keys are fixed
		//if s.Nodes[0] == All {
		//s.Nodes = t.Nodes
		//} else if s.Nodes[0] == Leader {
		////s.Nodes[0] = LeaderName
		//}

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
