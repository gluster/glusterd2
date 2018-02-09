package sunrpc

import (
	"bytes"
	"io"
	"log"
	"net/rpc"

	"github.com/rasky/go-xdr/xdr2"
)

type serverCodec struct {
	conn         io.ReadWriteCloser
	closed       bool
	notifyClose  chan<- io.ReadWriteCloser
	recordReader io.Reader
}

// NewServerCodec returns a new rpc.ServerCodec using Sun RPC on conn.
// If a non-nil channel is passed as second argument, the conn is sent on
// that channel when Close() is called on conn.
func NewServerCodec(conn io.ReadWriteCloser, notifyClose chan<- io.ReadWriteCloser) rpc.ServerCodec {
	return &serverCodec{conn: conn, notifyClose: notifyClose}
}

func (c *serverCodec) ReadRequestHeader(req *rpc.Request) error {
	// NOTE:
	// Errors returned by this function aren't relayed back to the client
	// as WriteResponse() isn't called. The net/rpc package will call
	// c.Close() when this function returns an error.

	// Read entire RPC message from network
	record, err := ReadFullRecord(c.conn)
	if err != nil {
		if err != io.EOF {
			log.Println(err)
		}
		return err
	}

	c.recordReader = bytes.NewReader(record)

	// Unmarshall RPC message
	var call RPCMsg
	_, err = xdr.Unmarshal(c.recordReader, &call)
	if err != nil {
		log.Println(err)
		return err
	}

	if call.Type != Call {
		log.Println(ErrInvalidRPCMessageType)
		return ErrInvalidRPCMessageType
	}

	// Set req.Seq and req.ServiceMethod
	req.Seq = uint64(call.Xid)
	procedureID := ProcedureID{call.CBody.Program, call.CBody.Version, call.CBody.Procedure}
	procedureName, ok := GetProcedureName(procedureID)
	if ok {
		req.ServiceMethod = procedureName
	} else {
		// Due to our simpler map implementation, we cannot distinguish
		// between ErrProgUnavail and ErrProcUnavail
		log.Printf("%s: %+v\n", ErrProcUnavail, procedureID)
		return ErrProcUnavail
	}

	return nil
}

func (c *serverCodec) ReadRequestBody(funcArgs interface{}) error {

	if funcArgs == nil {
		return nil
	}

	if _, err := xdr.Unmarshal(c.recordReader, &funcArgs); err != nil {
		c.Close()
		return err
	}

	return nil
}

func (c *serverCodec) WriteResponse(resp *rpc.Response, result interface{}) error {

	if resp.Error != "" {
		// The remote function returned error (shouldn't really happen)
		log.Println(resp.Error)
	}

	var buf bytes.Buffer

	reply := RPCMsg{
		Xid:  uint32(resp.Seq),
		Type: Reply,
		RBody: ReplyBody{
			Stat: MsgAccepted,
			Areply: AcceptedReply{
				Stat: Success,
			},
		},
	}

	if _, err := xdr.Marshal(&buf, reply); err != nil {
		c.Close()
		return err
	}

	// Marshal and fill procedure-specific reply into the buffer
	if _, err := xdr.Marshal(&buf, result); err != nil {
		c.Close()
		return err
	}

	// Write buffer contents to network
	if _, err := WriteFullRecord(c.conn, buf.Bytes()); err != nil {
		c.Close()
		return err
	}

	return nil
}

func (c *serverCodec) Close() error {
	if c.closed {
		return nil
	}

	err := c.conn.Close()
	if err == nil {
		c.closed = true
		if c.notifyClose != nil {
			c.notifyClose <- c.conn
		}
	}

	return err
}
