package transaction

import (
	"github.com/gluster/glusterd2/gdctx"

	log "github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
)

// SimpleTxn is transaction with fixed stage, commit and store steps
type SimpleTxn struct {
	// Ctx is the transaction context
	Ctx *TxnCtx
	// Nodes are the nodes where the stage and commit functions are performed
	Nodes []uuid.UUID
	// LockKey is the key to be locked
	LockKey string

	// Stage is the registered name of the staging StepFunc
	// Stage function verifies if the node can perform an operation.
	Stage string
	// Commit is the registered name of the commit StepFunc
	// Commit performs the operation on the a node
	Commit string
	// Store is the registered name of the store StepFunc
	// Store stores the results of an operation. This will only be run on the leader
	Store string
	// Rollback is the registered name of the rollback StepFunc
	// Rollback rolls back any changes done by Commit
	Rollback string
	//LogFields will be set in the transaction context
	LogFields *log.Fields
}

// NewTxn creates and returns a Txn using SimpleTxn as a template
func (s *SimpleTxn) NewTxn() (*Txn, error) {
	var simple *Txn
	if s.LogFields == nil {
		simple = NewTxn()
	} else {
		simple = NewTxnWithLoggingContext(*s.LogFields)
	}
	simple.Nodes = s.Nodes
	simple.Steps = make([]*Step, 5)

	lockstep, unlockstep, err := CreateLockSteps(s.LockKey)
	if err != nil {
		simple.Cleanup()
		return nil, err
	}

	stagestep := &Step{
		DoFunc:   s.Stage,
		UndoFunc: "",
		Nodes:    s.Nodes,
	}
	commitstep := &Step{
		DoFunc:   s.Commit,
		UndoFunc: s.Rollback,
		Nodes:    s.Nodes,
	}
	storestep := &Step{
		DoFunc:   s.Store,
		UndoFunc: "",
		Nodes:    []uuid.UUID{gdctx.MyUUID},
	}

	simple.Steps[0] = lockstep
	simple.Steps[1] = stagestep
	simple.Steps[2] = commitstep
	simple.Steps[3] = storestep
	simple.Steps[4] = unlockstep

	return simple, nil
}

// Do runs the SimpleTxn on the cluster
func (s *SimpleTxn) Do() (TxnCtx, error) {
	t, err := s.NewTxn()
	if err != nil {
		return nil, err
	}

	return t.Do()
}
