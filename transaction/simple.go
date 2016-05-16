package transaction

import (
	"github.com/gluster/glusterd2/context"
)

// SimpleTxn is transaction with fixed stage, commit and store steps
type SimpleTxn struct {
	// Name holds the command
	Name string

	// Ctx is the transaction context
	Ctx *context.Context
	// Nodes are the nodes where the stage and commit functions are performed
	Nodes []string

	// Stage function verifies if the node can perform an operation.
	Stage StepFunc
	// Commit performs the operation on the a node
	Commit StepFunc
	// Store stores the results of an operation. This will only be run on the leader
	Store StepFunc
	// Rollback rollsback any changes done by Commit
	Rollback StepFunc
}

// NewSimpleTxn returns creates and returns a Txn using e Simple transaction template
func NewSimpleTxn(c *context.Context, name string, nodes []string, stage, commit, store, rollback StepFunc) (*Txn, error) {
	simple := Txn{
		Ctx:   c,
		Name:  name,
		Nodes: nodes,
		Steps: make([]*Step, 5), //A simple transaction has just 5 steps
	}

	lockstep, unlockstep, err := CreateLockUnlockSteps()
	if err != nil {
		return nil, err
	}

	stagestep := &Step{
		DoFunc:   stage,
		UndoFunc: nil,
		Nodes:    nodes,
	}
	commitstep := &Step{
		DoFunc:   commit,
		UndoFunc: rollback,
		Nodes:    nodes,
	}
	storestep := &Step{
		DoFunc:   store,
		UndoFunc: nil,
		Nodes:    []string{Leader},
	}

	simple.Steps[0] = lockstep
	simple.Steps[1] = stagestep
	simple.Steps[2] = commitstep
	simple.Steps[3] = storestep
	simple.Steps[4] = unlockstep

	return &simple, nil
}
