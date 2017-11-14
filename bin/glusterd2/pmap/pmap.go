package pmap

import (
	"encoding/json"
	"expvar"
	"fmt"
	"net"
	"sync"
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

			if isPortFree(p) {
				registry.Ports[p].Type = GfPmapPortLeased
				port = p
				break
			} else {
				// We may have an opportunity here to change
				// the port's status from GfPmapPortFree to
				// GfPmapPortForeign. Passing on for now...
			}
		}
	}

	if port > registry.LastAlloc {
		registry.LastAlloc = port
	}

	return port
}

// AssignPort allocates and returns an available port. Optionally, if an
// oldport specified for the brickpath, stale ports for the brickpath will
// be cleaned up
func AssignPort(oldport int, brickpath string) int {
	if oldport != 0 {
		registryRemove(0, brickpath, GfPmapPortBrickserver, nil)
	}
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
