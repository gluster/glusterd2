package transaction

// Temporary declarations for step args and return.

// StepArg in the input to a Step
type StepArg interface{}

// StepRet is what the Step returns
type StepRet interface{}

// StepFunc is the function that is supposed to be run during a transaction step
type StepFunc func(StepArg) StepRet

const (
	//Leader is a constant string representing the leader node
	Leader = "leader"
	//All is a contant string representing all the nodes in a transaction
	All = "all"
)

// Step is a combination of a StepFunc and a list of nodes the step is supposed to be run on
//
// DoFunc performs does the action
// UndoFunc undoes anything done by DoFunc
// Nodes can have a single entry of either transaction.Leader or transaction.All, in which case the transaction will run on all Nodes in a transaction or just the leader
type Step struct {
	DoFunc   StepFunc
	UndoFunc StepFunc
	Nodes    []string
}

// do runs the DoFunc on the nodes
func (s *Step) do() error {
	for range s.Nodes {
		// RunStepFunconNode(s.DoFunc, n)
	}
	return nil
}

// undo runs the UndoFunc on the nodes
func (s *Step) undo() error {
	if s.UndoFunc != nil {
		for range s.Nodes {
			// RunStepFunconNode(s.UndoFunc, n)
		}
	}
	return nil
}
