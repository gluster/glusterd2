package bricksplanner

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/peer"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/utils"
	deviceapi "github.com/gluster/glusterd2/plugins/device/api"
)

var subvolPlanners = make(map[string]SubvolPlanner)

// SubvolPlanner represents planner interface for different subvol types
type SubvolPlanner interface {
	Init(*api.VolCreateReq, uint64)
	BricksCount() int
	BrickSize(int) uint64
	BrickType(int) string
}

// Vg represents Virtual Volume Group
type Vg struct {
	Name          string
	DeviceName    string
	PeerID        string
	Zone          string
	State         string
	AvailableSize uint64
	Used          bool
}

func getAvailableVgs(req *api.VolCreateReq) ([]Vg, error) {
	var vgs []Vg
	peers, err := peer.GetPeers()
	if err != nil {
		return nil, err
	}

	for _, p := range peers {
		// If Peer is not online, do not consider this device/peer
		if _, online := store.Store.IsNodeAlive(p.ID); !online {
			continue
		}

		peerzone, exists := p.Metadata["_zone"]
		if !exists || strings.TrimSpace(peerzone) == "" {
			peerzone = p.ID.String()
		}

		// If List of Peer IDs specified to limit choosing the bricks from
		if len(req.LimitPeers) > 0 && !utils.StringInSlice(p.ID.String(), req.LimitPeers) {
			continue
		}

		// If List of zones specified to limit choosing the bricks from
		if len(req.LimitZones) > 0 && !utils.StringInSlice(peerzone, req.LimitZones) {
			continue
		}

		// If Exclude List of Peer IDs specified
		if len(req.ExcludePeers) > 0 && utils.StringInSlice(p.ID.String(), req.ExcludePeers) {
			continue
		}

		// If Exclude List of Zones specified
		if len(req.ExcludeZones) > 0 && utils.StringInSlice(peerzone, req.ExcludeZones) {
			continue
		}

		devicesRaw, exists := p.Metadata["_devices"]
		if !exists {
			// No device registered for this peer
			continue
		}

		var deviceInfo []deviceapi.Info
		if err := json.Unmarshal([]byte(devicesRaw), &deviceInfo); err != nil {
			return nil, err
		}

		for _, d := range deviceInfo {
			// If Device is not enabled to be used for provisioning
			if d.State == deviceapi.DeviceDisabled {
				continue
			}

			vgs = append(vgs, Vg{
				DeviceName:    d.Name,
				Name:          d.VgName,
				PeerID:        p.ID.String(),
				Zone:          peerzone,
				State:         d.State,
				AvailableSize: d.AvailableSize,
				Used:          d.Used,
			})
		}
	}

	// Sort based on Free size available, this acts as rank/weightage each Vgs
	// during allocation. High priority will be given to the Vg with more FreeSize
	sort.Slice(vgs, func(i, j int) bool { return vgs[i].AvailableSize > vgs[j].AvailableSize })

	return vgs, nil
}
