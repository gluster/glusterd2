package transaction

import (
	"errors"
	"net"
	"net/rpc"

	"github.com/gluster/glusterd2/context"
	"github.com/gluster/glusterd2/peer"

	"github.com/kshlm/pbrpc/pbcodec"
	"github.com/pborman/uuid"
	config "github.com/spf13/viper"
)

// StepFunc is the function that is supposed to be run during a transaction step
type StepFunc func(*context.Context) error

//const (
////Leader is a constant string representing the leader node
//Leader = "leader"
////All is a contant string representing all the nodes in a transaction
//All = "all"
//)
// XXX: Because Nodes are now uuid.UUID, string constants cannot be used in node lists
// TODO: Figure out an alternate method and re-enable. Or just remove it.

// Step is a combination of a StepFunc and a list of nodes the step is supposed to be run on
//
// DoFunc and UndoFunc are names of StepFuncs registered in the registry
// DoFunc performs does the action
// UndoFunc undoes anything done by DoFunc
type Step struct {
	DoFunc   string
	UndoFunc string
	Nodes    []uuid.UUID
}

var (
	ErrStepFuncNotFound = errors.New("StepFunc was not found")
)

// do runs the DoFunc on the nodes
func (s *Step) do(c *context.Context) error {
	return runStepFuncOnNodes(s.DoFunc, c, s.Nodes)
}

// undo runs the UndoFunc on the nodes
func (s *Step) undo(c *context.Context) error {
	if s.UndoFunc != "" {
		return runStepFuncOnNodes(s.UndoFunc, c, s.Nodes)
	}
	return nil
}

func runStepFuncOnNodes(name string, c *context.Context, nodes []uuid.UUID) error {
	done := make(chan error)
	defer close(done)

	for i, node := range nodes {
		go runStepFuncOnNode(name, c, node, done)
	}

	// TODO: Need to properly aggregate results
	var err error
	for i >= 0 {
		err <- done
		i--
	}
	return err
}

func runStepFuncOnNode(name string, c *context.Context, node uuid.UUID, done chan<- error) {
	if node == context.MyUUID {
		done <- runStepFuncLocal(name, c)
	} else {
		done <- runStepFuncRemote(name, c, node)
	}
}

func runStepFuncLocal(name string, c *context.Context) error {
	c.Log.WithField("stepfunc", name).Debug("running step function")

	stepFunc, ok := GetStepFunc(name)
	if !ok {
		return ErrStepFuncNotFound
	}
	return stepFunc(c)
}

func runStepFuncRemote(name string, c *context.Context, node uuid.UUID) error {
	// TODO: I'm creating connections on demand. This should be changed so that
	// we have long term connections.
	p, err := peer.GetPeer(node.String())
	if err != nil {
		return err
	}

	var conn net.Conn
	port := config.GetString("rpcport")

	for _, addr := range p.Addresses {
		conn, err = net.Dial("tcp", addr+":"+port)
		if err == nil && conn != nil {
			break
		}
	}
	if conn == nil {
		return err
	}

	client := rpc.NewClientWithCodec(pbcodec.NewClientCodec(conn))
	defer client.Close()

}
