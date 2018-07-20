// Package volume contains some types associated with GlusterFS volumes that will be used in GlusterD
package volume

import (
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"
	"github.com/gluster/glusterd2/pkg/api"

	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
)

// VolState is the current status of a volume
type VolState uint16

const (
	// VolCreated should be set only for a volume that has been just created
	VolCreated VolState = iota
	// VolStarted should be set only for volumes that are running
	VolStarted
	// VolStopped should be set only for volumes that are not running, excluding newly created volumes
	VolStopped
)

var (
	absFilePath = filepath.Abs
	// NewBrickEntriesFunc creates the brick list
	NewBrickEntriesFunc = NewBrickEntries
)

// VolType is the status of the volume
//go:generate stringer -type=VolType
type VolType uint16

const (
	// Distribute is a plain distribute volume
	Distribute VolType = iota
	// Replicate is plain replicate volume
	Replicate
	// Disperse is a plain erasure coded volume
	Disperse
	// DistReplicate is a distribute-replicate volume
	DistReplicate
	// DistDisperse is a distribute-'erasure coded' volume
	DistDisperse
)

// SubvolType is the Type of the volume
type SubvolType uint16

const (
	// SubvolDistribute is a distribute sub volume
	SubvolDistribute SubvolType = iota
	// SubvolReplicate is a replicate sub volume
	SubvolReplicate
	// SubvolDisperse is a disperse sub volume
	SubvolDisperse
)

// Subvol represents a sub volume
type Subvol struct {
	ID              uuid.UUID
	Name            string
	Type            SubvolType
	Bricks          []brick.Brickinfo
	Subvols         []Subvol
	ReplicaCount    int
	ArbiterCount    int
	DisperseCount   int
	RedundancyCount int
}

// Volinfo repesents a volume
type Volinfo struct {
	ID        uuid.UUID
	Name      string
	VolfileID string
	Type      VolType
	Transport string
	DistCount int
	Options   map[string]string
	State     VolState
	Checksum  uint64
	Version   uint64
	Subvols   []Subvol
	Auth      VolAuth
	GraphMap  map[string]string
	Metadata  map[string]string
	SnapList  []string
}

// VolAuth represents username and password used by trusted/internal clients
type VolAuth struct {
	Username string
	Password string
}

// StringMap returns a map[string]string representation of Volinfo
func (v *Volinfo) StringMap() map[string]string {
	m := make(map[string]string)

	m["volume.id"] = v.ID.String()
	m["volume.name"] = v.Name
	m["volume.type"] = v.Type.String()
	m["volume.redundancy"] = "0"
	// TODO: Assumed First subvolume's redundancy count
	if len(v.Subvols) > 0 {
		m["volume.redundancy"] = strconv.Itoa(v.Subvols[0].RedundancyCount)
	}
	m["volume.transport"] = v.Transport
	m["volume.auth.username"] = v.Auth.Username
	m["volume.auth.password"] = v.Auth.Password

	return m
}

// NewBrickEntries creates the brick list
func NewBrickEntries(bricks []api.BrickReq, volName string, volID uuid.UUID) ([]brick.Brickinfo, error) {
	var brickInfos []brick.Brickinfo
	var binfo brick.Brickinfo

	for _, b := range bricks {
		u := uuid.Parse(b.PeerID)
		if u == nil {
			return nil, errors.New("invalid UUID specified as host for brick")
		}

		p, e := peer.GetPeerF(b.PeerID)
		if e != nil {
			return nil, e
		}

		binfo.PeerID = u
		// TODO: Have a better way to select peer address here
		binfo.Hostname, _, _ = net.SplitHostPort(p.PeerAddresses[0])

		binfo.Path, e = absFilePath(b.Path)
		if e != nil {
			log.Error("Failed to convert the brickpath to absolute path")
			return nil, e
		}

		switch b.Type {
		case "arbiter":
			binfo.Type = brick.Arbiter
		default:
			binfo.Type = brick.Brick
		}

		binfo.VolumeName = volName
		binfo.VolumeID = volID
		binfo.ID = uuid.NewRandom()

		// Auto provisioned bricks
		if b.VgName != "" && b.LvName != "" {
			binfo.MountInfo = brick.MountInfo{
				Mountdir:   b.Mountdir,
				DevicePath: b.DevicePath,
				FsType:     b.FsType,
				MntOpts:    b.MntOpts,
			}
		}

		brickInfos = append(brickInfos, binfo)
	}
	return brickInfos, nil
}

func (v *Volinfo) String() string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}

	var out bytes.Buffer
	json.Indent(&out, b, "", "\t")
	return out.String()
}

func (v *Volinfo) getBricks(onlyLocal bool) []brick.Brickinfo {
	var bricks []brick.Brickinfo

	for _, subvol := range v.Subvols {
		for _, b := range subvol.Bricks {
			if onlyLocal && !uuid.Equal(b.PeerID, gdctx.MyUUID) {
				continue
			}
			bricks = append(bricks, b)
		}
	}
	return bricks
}

// GetBricks returns a list of Bricks
func (v *Volinfo) GetBricks() []brick.Brickinfo {
	return v.getBricks(false)
}

// GetLocalBricks returns a list of local Bricks
func (v *Volinfo) GetLocalBricks() []brick.Brickinfo {
	return v.getBricks(true)
}

// Nodes returns the a list of nodes on which this volume has bricks
func (v *Volinfo) Nodes() []uuid.UUID {
	var nodes []uuid.UUID

	// This shouldn't be very inefficient for small slices.
	var present bool
	for _, b := range v.GetBricks() {
		// Add node to the slice only if it isn't present already
		present = false
		for _, n := range nodes {
			if uuid.Equal(b.PeerID, n) == true {
				present = true
				break
			}
		}

		if present == false {
			nodes = append(nodes, b.PeerID)
		}
	}

	return nodes
}

// Peers returns the a list of Peer objects on which this volume has bricks
func (v *Volinfo) Peers() []*peer.Peer {

	allPeers, err := peer.GetPeers()
	if err != nil {
		return nil
	}

	pDict := make(map[string]*peer.Peer, len(allPeers))
	for i := range allPeers {
		pDict[allPeers[i].ID.String()] = allPeers[i]
	}

	resultDict := make(map[string]*peer.Peer)
	for _, b := range v.GetBricks() {
		resultDict[b.PeerID.String()] = pDict[b.PeerID.String()]
	}

	var peers []*peer.Peer
	for _, v := range resultDict {
		peers = append(peers, v)
	}

	return peers
}

//SubvolTypeToString converts VolType to corresponding string
func SubvolTypeToString(subvolType SubvolType) string {
	switch subvolType {
	case SubvolReplicate:
		return "replicate"
	case SubvolDisperse:
		return "disperse"
	default:
		return "distribute"
	}
}

// MetadataSize returns the size of the volume metadata in Volume info
func (v *Volinfo) MetadataSize() int {
	size := 0
	for key, value := range v.Metadata {
		if !strings.HasPrefix(key, "_") {
			size = size + len(key) + len(value)
		}
	}
	return size
}

// GetLocalBricks returns a list of local Bricks
func (sv *Subvol) GetLocalBricks() []brick.Brickinfo {
	var bricks []brick.Brickinfo

	for _, b := range sv.Bricks {
		if !uuid.Equal(b.PeerID, gdctx.MyUUID) {
			continue
		}
		bricks = append(bricks, b)
	}
	return bricks
}
