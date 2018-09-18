package pmap

import (
	"expvar"
	"fmt"
	"net"
	"sync"
)

// IANA dynamic and/or private ports range
const (
	PortMin = 49152
	PortMax = 65535
)

// PortState represents the type of state of the port
type PortState int32

// List of states an individual port can be in
const (
	PortFree PortState = iota
	PortForeign
	PortLeased
	PortInUse
)

type portStatus struct {
	Port   int             `json:"port"`
	State  PortState       `json:"state"`
	Bricks map[string]bool `json:"bricks"`
}

type pmapRegistry struct {
	sync.RWMutex
	basePort  int
	lastAlloc int
	ports     [PortMax + 1]portStatus
	bricks    map[string]int   // map from brick path to port number
	conns     map[net.Conn]int // map from connection to port number

	portLockFds map[int]int
}

func (r *pmapRegistry) init() {
	for i := r.basePort; i <= PortMax; i++ {
		// TODO: There is a race in this check when there are multiple
		// glusterd2 instances running on the same machine.
		if isPortFree(i) {
			r.ports[i].State = PortFree
		} else {
			r.ports[i].State = PortForeign
		}
		r.ports[i].Port = i
	}
}

// Allocate finds a free port and returns the port number. The
// allocated free port will be in leased state until the brick
// process it was leased to sends a SIGN IN request to glusterd2.
func (r *pmapRegistry) Allocate(brickpath string) (int, error) {
	r.Lock()
	defer r.Unlock()

	var port int
	for p := r.basePort; p <= PortMax; p++ {
		if r.ports[p].State == PortFree || r.ports[p].State == PortForeign {
			// check if a port is free holding a lock on the port's "file"
			if r.TryPortLock(p) {
				if isPortFree(p) {
					r.ports[p].State = PortLeased
					if r.ports[p].Bricks == nil {
						r.ports[p].Bricks = make(map[string]bool)
					}
					r.ports[p].Bricks[brickpath] = false
					r.bricks[brickpath] = p
					port = p
					// keep port file locked if port was
					// found to be free and we leased it to
					// a brick process
					break
				} else {
					r.ports[p].State = PortForeign
					r.PortUnlock(p)
				}
			}
		}
	}

	if port == 0 {
		return -1, fmt.Errorf("registry.Allocate(): we ran out of free ports")
	}

	if port > r.lastAlloc {
		r.lastAlloc = port
	}

	return port, nil
}

// Update marks the port used by a brick with a specified state. This
// is called when a brick signs in.
func (r *pmapRegistry) Update(port int, brickpath string, conn net.Conn) error {

	if port < 0 || port > PortMax {
		return fmt.Errorf("registry.Update(): invalid port %d", port)
	}

	r.Lock()
	defer r.Unlock()

	// It's possible that multiple bricks are multiplexed onto a
	// single conn, the conn passed to this function may not be the
	// same as before and that can happen. We only store the latest
	// one.
	r.conns[conn] = port
	r.ports[port].State = PortInUse
	if r.ports[port].Bricks == nil {
		// can happen on glusterd2 restarts where we only get a SIGN IN
		r.ports[port].Bricks = make(map[string]bool)
	}
	r.ports[port].Bricks[brickpath] = true
	r.bricks[brickpath] = port

	if r.lastAlloc < port {
		r.lastAlloc = port
	}

	return nil
}

// SearchByBrickPath returns the port number used by the brick specified
// by the brick path provided.
func (r *pmapRegistry) SearchByBrickPath(brickpath string) (int, error) {

	if brickpath == "" {
		return -1, fmt.Errorf("SearchByBrickPath: brick path cannot be empty")
	}

	r.RLock()
	defer r.RUnlock()

	if port, ok := r.bricks[brickpath]; ok {
		return port, nil
	}

	return -1, fmt.Errorf("SearchByBrickPath: port for brick %s not found", brickpath)
}

// RemoveByConn deletes port map entry by brick process's TCP connection.
// There will be only one TCP connection per brick process, regardless of
// number of bricks in the process.
func (r *pmapRegistry) RemovePortByConn(conn net.Conn) error {

	if conn == nil {
		return fmt.Errorf("RemovePortByConn(): conn passed is nil")
	}

	r.Lock()
	defer r.Unlock()

	port, ok := r.conns[conn]
	if !ok {
		// this can happen in many cases:
		// * conn isn't a brick
		// * brick disconnects prior to SIGN IN
		return nil
	}

	delete(r.conns, conn)

	r.ports[port].State = PortFree
	r.PortUnlock(port)

	for brick := range r.ports[port].Bricks {
		delete(r.bricks, brick)
	}
	r.ports[port].Bricks = make(map[string]bool)

	return nil
}

// Remove deletes portmap entry of a single brick from the portmap registry.
// This is called when a brick process sends a SIGN OUT request to glusterd2
// during graceful shutdown.
func (r *pmapRegistry) Remove(port int, brickpath string, conn net.Conn) error {

	if port < 0 || port > PortMax {
		return fmt.Errorf("registry.Remove(): invalid port %d", port)
	}

	r.Lock()
	defer r.Unlock()

	delete(r.bricks, brickpath)
	if _, ok := r.ports[port].Bricks[brickpath]; ok {
		delete(r.ports[port].Bricks, brickpath)
	} else {
		return fmt.Errorf("registry.Remove(): invalid port %d and/or brick %s",
			port, brickpath)
	}

	// update connection object even on sign out
	r.conns[conn] = port
	return nil
}

var registry *pmapRegistry

// Init initializes the portmap registry. This is to be called when glusterd2
// server starts,
func Init() {

	if registry != nil {
		panic("registry is not nil: this shouldn't happen")
	}

	registry = &pmapRegistry{
		basePort:    PortMin,
		bricks:      make(map[string]int),
		conns:       make(map[net.Conn]int),
		portLockFds: make(map[int]int),
	}
	registry.init()

	expvar.Publish("pmap", registry)
}
