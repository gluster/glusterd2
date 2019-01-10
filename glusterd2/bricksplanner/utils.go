package bricksplanner

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/brick"
	"github.com/gluster/glusterd2/glusterd2/peer"
	"github.com/gluster/glusterd2/glusterd2/store"
	"github.com/gluster/glusterd2/glusterd2/volume"
	"github.com/gluster/glusterd2/pkg/api"
	"github.com/gluster/glusterd2/pkg/lvmutils"
	"github.com/gluster/glusterd2/pkg/utils"
	deviceapi "github.com/gluster/glusterd2/plugins/device/api"
	"github.com/gluster/glusterd2/plugins/device/deviceutils"
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
	Device        string
	PeerID        string
	Zone          string
	State         string
	AvailableSize uint64
	Used          bool
}

// GetAvailableVgs returns VG list that can be used to create bricks
func GetAvailableVgs(req *api.VolCreateReq) ([]Vg, error) {
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

		deviceInfo, err := deviceutils.GetDevices(p.ID.String())
		if err != nil {
			return nil, err
		}

		if len(deviceInfo) == 0 {
			// No device registered for this peer
			continue
		}

		for _, d := range deviceInfo {
			// If Device is not enabled to be used for provisioning
			if d.State == deviceapi.DeviceDisabled {
				continue
			}

			// If Provisioner type does not match the requested provisioner type
			if d.ProvisionerType != req.ProvisionerType {
				continue
			}

			vgs = append(vgs, Vg{
				Device:        d.Device,
				Name:          d.VgName(),
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

// GetNewBrick creates a new brick request for the new brick.
func GetNewBrick(availableVgs []Vg, brickInfo brick.Brickstatus, vol *volume.Volinfo, subVolIndex, brickIndex int) api.BrickReq {
	var newBrick api.BrickReq
	brickSize := brickInfo.Size.Capacity
	lvName := fmt.Sprintf("brick_%s_s%d_b%d", vol.Name, subVolIndex, brickIndex)
	brickTpSize := uint64(float64(brickSize) * vol.SnapshotReserveFactor)
	brickTpSize = lvmutils.NormalizeSize(brickTpSize)
	tpmsize := lvmutils.GetPoolMetadataSize(brickTpSize)
	for _, vg := range availableVgs {
		if vg.AvailableSize >= brickTpSize {

			newBrick = api.BrickReq{
				Type:           "brick",
				Path:           brickInfo.Info.Path,
				BrickDirSuffix: "/brick",
				TpName:         fmt.Sprintf("tp_%s_s%d_b%d", vol.Name, subVolIndex, brickIndex),
				LvName:         lvName,
				Size:           brickSize,
				TpSize:         brickTpSize,
				TpMetadataSize: tpmsize,
				FsType:         "xfs",
				MntOpts:        "rw,inode64,noatime,nouuid",
				PeerID:         vg.PeerID,
				VgName:         vg.Name,
				DevicePath:     "/dev/" + vg.Name + "/" + lvName,
				RootDevice:     vg.Device,
				TotalSize:      brickTpSize + tpmsize,
			}
			if vol.ProvisionerType == api.ProvisionerTypeLoop {
				newBrick.DevicePath = vg.Device + "/" + newBrick.TpName + "/" + newBrick.LvName + ".img"
			}
			vg.Used = true
			break
		}
	}
	return newBrick
}
