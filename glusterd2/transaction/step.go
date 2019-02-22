package transaction

import (
	"context"
	"errors"

	"github.com/pborman/uuid"
)

// StepFunc is the function that is supposed to be run during a transaction step
type StepFunc func(TxnCtx) error

// Step is a combination of a StepFunc and a list of nodes the step is supposed to be run on
//
// DoFunc and UndoFunc are names of StepFuncs registered in the registry
// DoFunc performs does the action
// UndoFunc undoes anything done by DoFunc
type Step struct {
	DoFunc   string
	UndoFunc string
	Nodes    []uuid.UUID
	Skip     bool
	Sync     bool
}

var (
	// ErrStepFuncNotFound is returned if the stepfunc isn't found.
	ErrStepFuncNotFound = errors.New("stepFunc was not found")
)

// RunStepFuncLocally runs a step func on local node
func RunStepFuncLocally(origCtx context.Context, stepName string, ctx TxnCtx) error {
	stepFunc, ok := getStepFunc(stepName)
	if !ok {
		return ErrStepFuncNotFound
	}

	if err := stepFunc(ctx); err != nil {
		return err
	}

	// if step function executes successfully, commit the
	// results to the store
	return ctx.Commit()
}
