package transaction

// SimpleTxn is transaction with fixed stage, commit and store steps
type SimpleTxn struct {
	// Nodes are the nodes where the stage and commit functions are performed
	Nodes []string

	// Stage function verifies if the node can perform an operation.
	Stage Step
	// Commit performs the operation on the a node
	Commit Step
	// Store stores the results of an operation. This will only be run on the leader
	Store Step
	// Rollback rollsback any changes done by Commit
	Rollback Step
}

// NewSimpleTxn returns creates and returns a Txn using e Simple transaction template
func NewSimpleTxn(nodes []string, stage, commit, store, rollback Step) Txn {
	simple := Txn{
		Nodes: nodes,
		Steps: make([]TxnStep, 3), //A simple transaction has just 3 steps
	}

	stagestep := TxnStep{
		Step:  stage,
		Nodes: []string{All},
	}
	commitstep := TxnStep{
		Step:  commit,
		Nodes: []string{All},
	}
	storestep := TxnStep{
		Step:  store,
		Nodes: []string{Leader},
	}

	simple.Steps[0] = stagestep
	simple.Steps[1] = commitstep
	simple.Steps[2] = storestep

	return simple
}

// Perform runs the SimpleTxn on the cluster
func (s *SimpleTxn) Perform() error {
	t := NewSimpleTxn(s.Nodes, s.Stage, s.Commit, s.Store, s.Rollback)

	return t.Perform()
}
