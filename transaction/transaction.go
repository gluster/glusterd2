// Package transaction implements a distributed transaction handling framework
package transaction

// GDTxnFw will be the GlusterD transaction framework
type GDTxnFw struct {
	// TODO: Add stuff as required
}

// Temporary declarations for step args and return.

// StepArg in the input to a Step
type StepArg interface{}

// StepRet is what the Step returns
type StepRet interface{}

// Step is the function that is supposed to be run during a transaction step
type Step func(StepArg) StepRet

const (
	//Leader is a constant string representing the leader node
	Leader = "leader"
	//All is a contant string representing all the nodes in a transaction
	All = "all"
)

// TxnStep is a combination of a Step function and a list of nodes the step is supposed to be run on
//
// TxnStep can have a single entry of either transaction.Leader or transaction.All, in which case the transaction will run on all Nodes or just the Leader
type TxnStep struct {
	Step  Step
	Nodes []string
}

// Do performs the step
func (s *TxnStep) Do() error {
	for _, n := range s.Nodes {
		// DoSteponNode(s.Step, n)
	}
	return nil
}

// Txn is a set of steps
//
// Nodes is a union of the all the TxnStep.Nodes
type Txn struct {
	Steps []TxnStep
	Nodes []string
}

// Perform runs the transaction on the cluster
func (t *Txn) Perform() error {
	//First verify all nodes are online
	for _, n := range t.Nodes {
		/*
			if !Online(n) {
				return error
			}
		*/
	}

	//Do the steps
	for _, s := range t.Steps {
		if s.Nodes[0] == All {
			s.Nodes = t.Nodes
		} else if s.Nodes[0] == Leader {
			//s.Nodes[0] = LeaderName
		}

		if e := s.Do(); e != nil {
			return e
		}
	}

	return nil
}
