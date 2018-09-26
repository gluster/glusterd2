package pmap

import (
	"encoding/json"
)

// The string and json support is to ensure that all the port information
// is dumped in the statedump response/output for debugging.

func (r *pmapRegistry) String() string {
	mb, _ := json.Marshal(r)
	return string(mb)
}

func (r *pmapRegistry) MarshalJSON() ([]byte, error) {

	r.RLock()
	defer r.RUnlock()

	var ports []portStatus
	for p := r.basePort; p <= r.lastAlloc; p++ {
		if r.ports[p].State == PortForeign || r.ports[p].State == PortFree {
			continue
		}
		ports = append(ports, r.ports[p])
	}

	return json.Marshal(ports)
}
