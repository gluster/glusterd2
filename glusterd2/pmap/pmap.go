package pmap

import (
	"encoding/json"
	"expvar"
	"fmt"
	"net"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"
)

const (
	gfIanaPrivPortsStart = 49152
	gfPortMax            = 65535
)

// PortType represents the type of state of the port
type PortType int32

// List of states an individual port can be in
const (
	GfPmapPortFree PortType = iota
	GfPmapPortForeign
	GfPmapPortLeased
	GfPmapPortNone
	GfPmapPortBrickserver
)

type portStatus struct {
	Port   int         `json:"port"`
	Type   PortType    `json:"state"`
	Bricks []string    `json:"bricks"`
	Xprt   interface{} `json:"-"`
}

type registryType struct {
	sync.RWMutex `json:"-"`
	BasePort     int                       `json:"base_port"`
	LastAlloc    int                       `json:"last_allocated_port,omitempty"`
	Ports        [gfPortMax + 1]portStatus `json:"ports,omitempty"`

	portLockFds map[int]int
}

func (r *registryType) String() string {
	mb, _ := json.Marshal(r)
	return string(mb)
}

func (r *registryType) MarshalJSON() ([]byte, error) {
	var ports []portStatus
	registry.RLock()
	for p := registry.BasePort; p <= registry.LastAlloc; p++ {
		if len(registry.Ports[p].Bricks) == 0 {
			continue
		}
		ports = append(ports, registry.Ports[p])
	}
	registry.RUnlock()

	type aliasType registryType
	return json.Marshal(&struct {
		*aliasType
		Ports []portStatus `json:"ports,omitempty"`
	}{
		aliasType: (*aliasType)(r),
		Ports:     ports,
	})
}

// This change synchronizes port allocation on a node/machine when multiple
// glusterd2 instances are running (e2e tests). This solution is to be
// treated as stopgap fix. A long term and more robust solution is to let
// the bricks pick and bind on their own.

func (r *registryType) TryPortLock(port int) bool {

	if _, ok := r.portLockFds[port]; ok {
		// we already have a lock
		return false
	}

	portLockFile := fmt.Sprintf("/var/run/glusterd2_pmap_port_%d.lock", port)
	fd, err := syscall.Open(portLockFile,
		syscall.O_CREAT|syscall.O_WRONLY|syscall.O_CLOEXEC, 0666)
	if err != nil {
		return false
	}

	err = syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB)
	switch err {
	case nil:
		// keep the fd open if we get the lock, close otherwise
		r.portLockFds[port] = fd
		log.WithField("lock-file", portLockFile).Debug(
			"Obtained lock on pmap port lock file.")
		return true
	case syscall.EWOULDBLOCK:
		log.WithField("lock-file", portLockFile).Debug(
			"Failed to obtain lock on pmap port lock file.")
	}

	syscall.Close(fd)
	return false
}

func (r *registryType) PortUnlock(port int) {
	fd, ok := r.portLockFds[port]
	if !ok {
		return
	}
	syscall.Flock(fd, syscall.LOCK_UN)
	syscall.Close(fd)
	delete(r.portLockFds, port)
	if registry.Ports[port].Type == GfPmapPortFree || registry.Ports[port].Type == GfPmapPortForeign {
		portLockFile := fmt.Sprintf("/var/run/glusterd2_pmap_port_%d.lock", port)
		syscall.Unlink(portLockFile)
	}
}

var registry = new(registryType)

func isPortFree(port int) bool {
	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func registrySearchByXprt(xprt interface{}, ptype PortType) int {
	registry.RLock()
	defer registry.RUnlock()

	var port int
	for p := registry.LastAlloc; p >= registry.BasePort; p-- {
		if registry.Ports[p].Xprt == nil {
			continue
		}
		if (registry.Ports[p].Xprt == xprt) && (registry.Ports[p].Type == ptype) {
			port = p
			break
		}
	}
	return port
}

func stringInSlice(query string, list []string) bool {
	for _, s := range list {
		if s == query {
			return true
		}
	}
	return false
}

// RegistrySearch searches for a brick process in the portmap registry and
// returns the port assigned to it.
// NOTE: Unlike glusterd1's implementation, the search here is not overloaded
// with delete operation. This is intentionally kept simple
func RegistrySearch(brickname string, ptype PortType) int {
	registry.RLock()
	defer registry.RUnlock()

	for p := registry.LastAlloc; p >= registry.BasePort; p-- {

		if len(registry.Ports[p].Bricks) == 0 || registry.Ports[p].Type != ptype {
			continue
		}

		if stringInSlice(brickname, registry.Ports[p].Bricks) {
			return p
		}
	}

	return 0
}

func registryAlloc(recheckForeign bool) int {
	registry.Lock()
	defer registry.Unlock()

	var port int
	for p := registry.BasePort; p <= gfPortMax; p++ {
		if registry.Ports[p].Type == GfPmapPortFree ||
			(recheckForeign && registry.Ports[p].Type == GfPmapPortForeign) {

			if registry.TryPortLock(p) {
				if isPortFree(p) {
					registry.Ports[p].Type = GfPmapPortLeased
					port = p
					// keep port file locked if port was
					// found to be free and we leased it to
					// a brick
					break
				} else {
					registry.Ports[p].Type = GfPmapPortForeign
					registry.PortUnlock(p)
				}
			}
		}
	}

	if port > registry.LastAlloc {
		registry.LastAlloc = port
	}

	return port
}

// AssignPort allocates and returns an available port. It also cleans up old
// stale ports.
func AssignPort(oldport int, brickpath string) int {
	// cleanup stale assigned and leased ports
	registryRemove(oldport, brickpath, GfPmapPortBrickserver, nil)
	registryRemove(oldport, brickpath, GfPmapPortLeased, nil)
	return registryAlloc(true)
}

func registryBind(port int, brickname string, ptype PortType, xprt interface{}) {

	if port > gfPortMax {
		return
	}

	registry.Lock()
	defer registry.Unlock()

	registry.Ports[port].Type = ptype
	registry.Ports[port].Bricks = append(registry.Ports[port].Bricks, brickname)
	registry.Ports[port].Xprt = xprt

	if registry.LastAlloc < port {
		registry.LastAlloc = port
	}
}

// opposite of append(), fast but doesn't maintain order
func deleteFromSlice(list []string, query string) []string {

	var found bool
	var pos int
	for i, s := range list {
		if s == query {
			pos = i
			found = true
			break
		}
	}

	if found {
		// swap i'th element with the last element
		list[len(list)-1], list[pos] = list[pos], list[len(list)-1]
		return list[:len(list)-1]
	}

	return list
}

func doRemove(port int, brickname string, xprt interface{}) {
	// TODO: This code below needs some more careful attention and actual
	// testing; especially around presence/absence of Xprt in case of
	// multiplexed bricks, tierd and snapd - all of which seem to use the
	// pmap service.

	registry.Lock()
	defer registry.Unlock()
	if len(registry.Ports[port].Bricks) == 1 {
		// Bricks aren't multiplexed over the same port
		// clear the bricknames array and reset other fields
		registry.Ports[port].Bricks = registry.Ports[port].Bricks[:0]
		registry.Ports[port].Type = GfPmapPortFree
		registry.Ports[port].Xprt = nil
	} else {
		// Bricks are multiplexed, only remove the brick entry
		registry.Ports[port].Bricks =
			deleteFromSlice(registry.Ports[port].Bricks, brickname)
		if (xprt != nil) && (xprt == registry.Ports[port].Xprt) {
			registry.Ports[port].Xprt = nil
		}
	}
	registry.PortUnlock(port)
}

func registryRemove(port int, brickname string, ptype PortType, xprt interface{}) {
	if port > 0 {
		if port > gfPortMax {
			return
		}
	}

	if brickname != "" {
		port = RegistrySearch(brickname, ptype)
		if port != 0 {
			doRemove(port, brickname, xprt)
			return
		}
	}

	if xprt != nil {
		port = registrySearchByXprt(xprt, ptype)
		if port != 0 {
			doRemove(port, brickname, xprt)
		}
	}
}

var registryInit sync.Once

func initRegistry() {
	registry.Lock()
	defer registry.Unlock()

	registry.portLockFds = make(map[int]int)

	// TODO: When a config option by the name 'base-port'
	// becomes available, use that.
	registry.BasePort = gfIanaPrivPortsStart

	for i := registry.BasePort; i <= gfPortMax; i++ {
		if isPortFree(i) {
			registry.Ports[i].Type = GfPmapPortFree
		} else {
			registry.Ports[i].Type = GfPmapPortForeign
		}
		registry.Ports[i].Port = i
	}
}

func init() {
	registryInit.Do(initRegistry)
	expvar.Publish("pmap", registry)
}
