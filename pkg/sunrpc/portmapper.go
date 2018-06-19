package sunrpc

import (
	"errors"
	"net"
	"net/rpc"
	"strconv"
	"sync"
)

const (
	pmapPort                 = 111
	portmapperProgramNumber  = 100000
	portmapperProgramVersion = 2
)

// Protocol is a type representing the protocol (TCP or UDP) over which the
// program/server being registered listens on.
type Protocol uint32

const (
	// IPProtoTCP is the protocol number for TCP/IP
	IPProtoTCP Protocol = 6
	// IPProtoUDP is the protocol number for UDP/IP
	IPProtoUDP Protocol = 17
)

var defaultAddress = "127.0.0.1:" + strconv.Itoa(pmapPort)

// PortMapping is a mapping between (program, version, protocol) to port number
type PortMapping struct {
	Program  uint32
	Version  uint32
	Protocol uint32
	Port     uint32
}

var registryInit sync.Once

func initRegistry() {

	procedureID := ProcedureID{
		ProgramNumber:  portmapperProgramNumber,
		ProgramVersion: portmapperProgramVersion,
	}

	// This is ordered as per procedure number
	remoteProcedures := [6]string{
		"Pmap.ProcNull", "Pmap.ProcSet", "Pmap.ProcUnset",
		"Pmap.ProcGetPort", "Pmap.ProcDump", "Pmap.ProcCallIt"}

	for id, procName := range remoteProcedures {
		procedureID.ProcedureNumber = uint32(id)
		_ = RegisterProcedure(Procedure{procedureID, procName}, true)
	}
}

func initPmapClient(host string) *rpc.Client {
	if host == "" {
		host = defaultAddress
	}

	registryInit.Do(initRegistry)

	conn, err := net.Dial("tcp", host)
	if err != nil {
		return nil
	}

	return rpc.NewClientWithCodec(NewClientCodec(conn, nil))
}

// PmapSet creates port mapping of the program specified. It return true on
// success and false otherwise.
func PmapSet(programNumber, programVersion uint32, protocol Protocol, port uint32) (bool, error) {

	var result bool

	client := initPmapClient("")
	if client == nil {
		return result, errors.New("could not create pmap client")
	}
	defer client.Close()

	mapping := &PortMapping{
		Program:  programNumber,
		Version:  programVersion,
		Protocol: uint32(protocol),
		Port:     port,
	}

	err := client.Call("Pmap.ProcSet", mapping, &result)
	return result, err
}

// PmapUnset will unregister the program specified. It returns true on success
// and false otherwise.
func PmapUnset(programNumber, programVersion uint32) (bool, error) {

	var result bool

	client := initPmapClient("")
	if client == nil {
		return result, errors.New("could not create pmap client")
	}
	defer client.Close()

	mapping := &PortMapping{
		Program: programNumber,
		Version: programVersion,
	}

	err := client.Call("Pmap.ProcUnset", mapping, &result)
	return result, err
}

// PmapGetPort returns the port number on which the program specified is
// awaiting call requests. If host is empty string, localhost is used.
func PmapGetPort(host string, programNumber, programVersion uint32, protocol Protocol) (uint32, error) {

	var port uint32

	client := initPmapClient(host)
	if client == nil {
		return port, errors.New("could not create pmap client")
	}
	defer client.Close()

	mapping := &PortMapping{
		Program:  programNumber,
		Version:  programVersion,
		Protocol: uint32(protocol),
	}

	err := client.Call("Pmap.ProcGetPort", mapping, &port)
	return port, err
}

type portMappingList struct {
	Map  PortMapping
	Next *portMappingList `xdr:"optional"`
}

type getMapsReply struct {
	Next *portMappingList `xdr:"optional"`
}

// PmapGetMaps returns a list of PortMapping entries present in portmapper's
// database. If host is empty string, localhost is used.
func PmapGetMaps(host string) ([]PortMapping, error) {

	var mappings []PortMapping
	var result getMapsReply

	client := initPmapClient(host)
	if client == nil {
		return nil, errors.New("could not create pmap client")
	}
	defer client.Close()

	err := client.Call("Pmap.ProcDump", nil, &result)
	if err != nil {
		return nil, err
	}

	if result.Next != nil {
		trav := result.Next
		for {
			entry := PortMapping(trav.Map)
			mappings = append(mappings, entry)
			trav = trav.Next
			if trav == nil {
				break
			}
		}
	}

	return mappings, nil
}
