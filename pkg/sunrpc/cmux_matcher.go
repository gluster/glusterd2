package sunrpc

import (
	"bytes"
	"encoding/binary"
	"io"
	"unsafe"

	"github.com/rasky/go-xdr/xdr2"
)

type rpcHeader struct {
	Xid        uint32
	MsgType    int32
	RPCVersion uint32
	Program    uint32
	Version    uint32
	Procedure  uint32
}

var minFragmentSize = unsafe.Sizeof(rpcHeader{})
var maxRPCReadSize = 4 + minFragmentSize // 28 bytes

// CmuxMatcher reads 28 bytes of the request to guess if the request is
// a Sun RPC Call. You can also match RPC requests targeted at specific
// program and version by passing variable params (hack for lack of function
// overloading)
func CmuxMatcher(progAndVersion ...uint32) func(io.Reader) bool {
	return func(reader io.Reader) bool {
		// read from connection
		buf := make([]byte, maxRPCReadSize)
		bytesRead, err := io.ReadFull(reader, buf)
		if err != nil || bytesRead != int(maxRPCReadSize) {
			return false
		}
		bufReader := bytes.NewReader(buf)

		// validate fragment size
		var fragmentHeader uint32
		err = binary.Read(bufReader, binary.BigEndian, &fragmentHeader)
		if err != nil {
			return false
		}
		fragmentSize := getFragmentSize(fragmentHeader)
		if fragmentSize < uint32(minFragmentSize) || fragmentSize > uint32(maxRecordFragmentSize) {
			return false
		}

		// validate RPC call
		var header rpcHeader
		bytesRead, err = xdr.Unmarshal(bufReader, &header)
		if err != nil || bytesRead != int(minFragmentSize) {
			return false
		}
		if header.MsgType != int32(Call) {
			return false
		}
		if header.RPCVersion != RPCProtocolVersion {
			return false
		}
		if header.Version == uint32(0) {
			return false
		}

		// match specific program number and version
		if len(progAndVersion) == 2 {
			if progAndVersion[0] != header.Program || progAndVersion[1] != header.Version {
				return false
			}
		}

		return true
	}
}
