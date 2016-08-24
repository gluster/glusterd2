// Package transaction implements a distributed transaction handling framework
package transaction

import (
	"github.com/gluster/glusterd2/context"
	"github.com/gluster/glusterd2/store"

	log "github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
)

const (
	txnPrefix = store.GlusterPrefix + "/transaction/"
)

func init() {
	context.RegisterStorePrefix(txnPrefix)
}

// Txn is a set of steps
//
// Nodes is a union of the all the TxnStep.Nodes
type Txn struct {
	ID    uuid.UUID
	Ctx   *Context
	Steps []*Step
	Nodes []uuid.UUID
}

// NewTxn returns an initialized Txn without any steps
func NewTxn() *Txn {
	t := new(Txn)
	t.ID = uuid.NewRandom()
	t.Ctx = NewLoggingContext(log.Fields{
		"txnid": t.ID.String(),
	})
	t.Ctx.Prefix = txnPrefix + t.ID.String()
	context.Store.InitPrefix(t.Ctx.Prefix)

	return t
}

// Do runs the transaction on the cluster
func (t *Txn) Do() (*context.Context, error) {
	if e := t.prepareTxn(); e != nil {
		return nil, e
	}


// Cleanup cleans the leftovers after a transaction ends
func (t *Txn) Cleanup() {
	context.Store.Delete(t.Ctx.Prefix)
}

// Do runs the transaction on the cluster
func (t *Txn) Do() (*Context, error) {
	t.Ctx.Log.Debug("Starting transaction")

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
			t.Ctx.Log.WithError(e).Error("Transaction failed, rolling back changes")
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
