// Package transaction implements a distributed transaction handling framework
package transaction

// GDTxnFw will be the GlusterD transaction framework
type GDTxnFw struct {
	// TODO: Add stuff as required
}

// Txn is a set of steps
//
// Nodes is a union of the all the TxnStep.Nodes
type Txn struct {
	Steps []Step
	Nodes []string
}

// Do runs the transaction on the cluster
func (t *Txn) Do() error {
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

		if e := s.do(); e != nil {
			t.undo(i)
		}
	}

	return nil
}

// undo undoes a transaction and will be automatically called by Perform if any step fails.
// The Steps are undone in the reverse order, from the failed step.
func (t *Txn) undo(n int) {
	for i := n; i >= 0; i-- {
		t.Steps[i].undo()
	}
}
