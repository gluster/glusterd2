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

	// Nodes is the union of the all the TxnStep.Nodes and is implicitly
	// set in Txn.Do(). This list is used to determine liveness of the
	// nodes before running the transaction steps.
	DontCheckAlive bool
	Nodes          []uuid.UUID

	DisableRollback bool
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

	for i, s := range t.Steps {
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
