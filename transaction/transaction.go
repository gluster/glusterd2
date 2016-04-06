// Package transaction implements a distributed transaction handling framework
package transaction

import (
	"github.com/gluster/glusterd2/context"

	log "github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
)

// Txn is a set of steps
//
// Nodes is a union of the all the TxnStep.Nodes
type Txn struct {
	Ctx   *context.Context
	Steps []*Step
	Nodes []string
}

// prepareTxn sets up some stuff required for the transaction
// like setting a transaction id
func (t *Txn) prepareTxn() error {
	t.Ctx = t.Ctx.NewLoggingContext(log.Fields{
		"txnid": uuid.NewRandom().String(),
	})

	return nil
}

// Do runs the transaction on the cluster
func (t *Txn) Do() error {
	if e := t.prepareTxn(); e != nil {
		return e
	}

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
		if s.Nodes[0] == All {
			s.Nodes = t.Nodes
		} else if s.Nodes[0] == Leader {
			//s.Nodes[0] = LeaderName
		}

		if e := s.do(t.Ctx); e != nil {
			t.Ctx.Log.WithError(e).Error("Transaction failed, rolling back changes")
			t.undo(i)
			return e
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
