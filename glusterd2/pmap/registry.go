package pmap

import (
	"encoding/json"
	"expvar"
	"fmt"
	"net"
	"sync"
)

// common ephemeral port range across IANA's range (49152 to 65535),
// linux defaults (32768 to 61000) and BSD defaults (1024 to 5000).
const (
	portMin = 1024
	portMax = 65535
)

type brickSet map[string]struct{}

type pmapRegistry struct {
	sync.RWMutex

	// map from brick path to port number
	// used to serve BrickByPort RPC request sent by clients during mount
	bricks map[string]int

	// map from connection to port number
	// used to process disconnections
	conns map[net.Conn]int

	// map from port number to list of bricks
	// used to process disconnections
	Ports map[int]brickSet `json:"ports,omitempty"`
}

func (r *pmapRegistry) String() string {
	b, err := json.Marshal(r)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

// Update marks the port used by a brick with a specified state. This
// is called when a brick signs in.
func (r *pmapRegistry) Update(port int, brickpath string, conn net.Conn) error {

	if port < portMin || port > portMax {
		return fmt.Errorf("registry.Update(): invalid port %d", port)
	}

	r.Lock()
	defer r.Unlock()

	r.bricks[brickpath] = port

	// It's possible that multiple bricks are multiplexed onto a
	// single conn, the conn passed to this function may not be the
	// same as before and that can happen. We only store the latest
	// one.
	r.conns[conn] = port

	if r.Ports[port] == nil {
		r.Ports[port] = make(map[string]struct{})
	}
	r.Ports[port][brickpath] = struct{}{}

	return nil
}

// SearchByBrickPath returns the port number used by the brick specified
// by the brick path provided. This is called when serving BrickByPort
// RPC request sent by the client during mount.
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

	for brick := range r.Ports[port] {
		delete(r.bricks, brick)
	}
	delete(r.Ports, port)

	return nil
}

// Remove deletes portmap entry of a single brick from the portmap registry.
// This is called when a brick process sends a SIGN OUT request to glusterd2
// during graceful shutdown.
func (r *pmapRegistry) Remove(port int, brickpath string, conn net.Conn) error {

	if port < portMin || port > portMax {
		return fmt.Errorf("registry.Remove(): invalid port %d", port)
	}

	r.Lock()
	defer r.Unlock()

	delete(r.bricks, brickpath)

	delete(r.Ports[port], brickpath)

	// update connection object even on sign out
	r.conns[conn] = port

	return nil
}

var registry *pmapRegistry

func init() {

	if registry != nil {
		panic("registry is not nil: this shouldn't happen")
	}

	registry = &pmapRegistry{
		Ports:  make(map[int]brickSet),
		bricks: make(map[string]int),
		conns:  make(map[net.Conn]int),
	}

	expvar.Publish("pmap", registry)
}
