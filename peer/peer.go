// Package peer implements the Peer type
package peer

import (
	"encoding/json"

	"github.com/pborman/uuid"
)

// Peer reperesents a GlusterD
type Peer struct {
	ID        uuid.UUID `json:"-"`
	IDstr     string    `json:"id"`
	Name      string    `json:"name"`
	Addresses []string  `json:"addresses"`
}

// UnmarshalJSON unmarshalls a JSON representation of the peer and returns a Peer
func UnmarshalJSON(j []byte) (*Peer, error) {
	p := &Peer{}

	err := p.UnmarshalJSON(j)

	return p, err
}

// UnmarshalJSON unmarshals a JSON representation of a peer into the given `p`
func (p *Peer) UnmarshalJSON(j []byte) error {
	if err := json.Unmarshal(j, p); err != nil {
		return err
	}

	p.ID = uuid.Parse(p.IDstr)

	return nil
}
