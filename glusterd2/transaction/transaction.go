// Package transaction implements a distributed transaction handling framework
package transaction

import (
	"context"
	"expvar"
	"fmt"

	"github.com/gluster/glusterd2/glusterd2/store"

	"github.com/coreos/etcd/clientv3"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

const (
	txnPrefix = store.GlusterPrefix + "transaction/"
)

var expTxn = expvar.NewMap("txn")

// Txn is a set of steps
type Txn struct {
	ID    uuid.UUID
	Ctx   TxnCtx
	Steps []*Step
	// Nodes should be a union of the all the TxnStep.Nodes and must be set
	// before calling Txn.Do(). This is currently only used to determine
	// liveness of the nodes before running the transaction steps.
	Nodes []uuid.UUID
}

// NewTxn returns an initialized Txn without any steps
func NewTxn(id string) *Txn {
	t := new(Txn)

	if t.ID = uuid.Parse(id); t.ID == nil {
		t.ID = uuid.NewRandom()
		log.WithField("reqid", t.ID.String()).Warn("Invalid UUID set as request ID. Generated new request ID")
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
