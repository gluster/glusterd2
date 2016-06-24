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
	Name  string
	Ctx   *context.Context
	Steps []*Step
	Nodes []string
}

// TxnSteps contains the bare minimum details to be maintained to capture the
// transaction steps
/*type TxnSteps struct {
	Name  string
	Steps []*Step
}*/

// Txns is a table of all the txn steps for the command
type Txns []*Txn

// TxnTable contains all the transaction details
var (
	TxnTable *Txns
)

// SetTxnSteps sets all the txns in the global TxnTable
func SetTxnSteps(txns *Txns) {
	TxnTable = txns
}

// RegisterTxn registers all the transaction steps
func RegisterTxn(name string, stage StepFunc, commit StepFunc, store StepFunc, rollback StepFunc) *Txn {
	t, err := NewSimpleTxn(nil, name, []string{All}, stage, commit, store, rollback)
	if err != nil {
		return nil
	}
	return t
}

// GetTxn retrieves the txn from the global txn table
func GetTxn(name string) *Txn {
	for _, t := range *TxnTable {
		if t.Name == name {
			return t
		}
	}
	return nil
}

// UpdateTxn retrieves the txn and updates it with ctx and nodes
func (t *Txn) UpdateTxn(ctx *context.Context, nodes []string) {
	t.Ctx = ctx
	t.Nodes = nodes
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
