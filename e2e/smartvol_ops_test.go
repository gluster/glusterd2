package e2e

import (
	"fmt"
	"syscall"
	"testing"

	"github.com/gluster/glusterd2/pkg/api"
	gutils "github.com/gluster/glusterd2/pkg/utils"
	deviceapi "github.com/gluster/glusterd2/plugins/device/api"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
)

func brickSizeTest(brickpath string, min uint64, max uint64) error {
	var fstat syscall.Statfs_t
	if err := syscall.Statfs(brickpath, &fstat); err != nil {
		return fmt.Errorf("unable to get size info of Brick(%s) %v", brickpath, err)
	}

	if &fstat != nil {
		value := uint64((fstat.Blocks * uint64(fstat.Bsize)) / (1024 * 1024))
		if value < min || value > max {
			return fmt.Errorf("Brick(%s) size mismatch, expected: %d-%d, got: %d", brickpath, min, max, value)
		}
		return nil
	}

	return fmt.Errorf("unable to get size info of Brick(%s)", brickpath)
}

func checkZeroLvs(r *require.Assertions) {
	checkZeroLvsWithRange(r, 1, 2)
}

func checkZeroLvsWithRange(r *require.Assertions, start, end int) {
	for i := start; i <= end; i++ {
		nlv, err := numberOfLvs(fmt.Sprintf("gluster-dev-gluster_loop%d", i))
		r.Nil(err)
		if err == nil {
			r.Equal(0, nlv)
		}
	}
}

func getPeerIDs(subvols []api.Subvol) []uuid.UUID {
	peers := make([]uuid.UUID, 0)
	for _, subvol := range subvols {
		for _, brick := range subvol.Bricks {
			peers = append(peers, brick.PeerID)
		}
	}
	return peers
}

// Replace brick test
func testReplaceBrick(t *testing.T) {
	r := require.New(t)
	smartvolname := formatVolName(t.Name())
	// create Distribute 3 Volume
	createReq := api.VolCreateReq{
		Name:            smartvolname,
		Size:            60 * gutils.MiB,
		DistributeCount: 2,
	}

	// Create distribute volume
	volinfo, err := client.VolumeCreate(createReq)
	r.Nil(err)

	r.Len(volinfo.Subvols, 2)
	r.Equal("Distribute", volinfo.Type.String())
	r.Len(volinfo.Subvols[0].Bricks, 1)
	r.Len(volinfo.Subvols[1].Bricks, 1)

	// Start volume
	err = client.VolumeStart(smartvolname, true)
	r.Nil(err)

	oldBrickPath := volinfo.Subvols[0].Bricks[0].Path
	oldBrickPeerID := volinfo.Subvols[0].Bricks[0].PeerID

	replaceBrickReq := api.ReplaceBrickReq{
		SrcPeerID:    oldBrickPeerID.String(),
		SrcBrickPath: oldBrickPath,
		Force:        true,
	}

	volInfo, err := client.ReplaceBrick(smartvolname, replaceBrickReq)
	r.Nil(err)
	r.NotEqual(volInfo.Subvols[0].Bricks[0].PeerID.String(), oldBrickPeerID)

	err = client.VolumeStop(smartvolname)
	r.Nil(err)
	r.Nil(client.VolumeDelete(smartvolname))

	g4, err := spawnGlusterd(t, "./config/4.toml", true)
	r.Nil(err)
	defer g4.Stop()
	r.True(g4.IsRunning())

	peerAddReq := api.PeerAddReq{
		Addresses: []string{g4.PeerAddress},
		Metadata: map[string]string{
			"owner": "gd4test",
		},
	}
	peerinfo, err := client.PeerAdd(peerAddReq)
	r.Nil(err)

	volinfo, err = client.VolumeCreate(createReq)
	r.Nil(err)

	// Start volume
	err = client.VolumeStart(smartvolname, true)
	r.Nil(err)

	oldBrickPath = volinfo.Subvols[0].Bricks[0].Path
	oldBrickPeerID = volinfo.Subvols[0].Bricks[0].PeerID
	excludePeer := []string{peerinfo.ID.String()}
	replaceBrickReq = api.ReplaceBrickReq{
		SrcPeerID:    oldBrickPeerID.String(),
		SrcBrickPath: oldBrickPath,
		ExcludePeers: excludePeer,
		Force:        true,
	}

	_, err = client.ReplaceBrick(smartvolname, replaceBrickReq)
	r.Nil(err)

	r.NotEqual(peerinfo.ID, oldBrickPeerID)

	err = client.VolumeStop(smartvolname)
	r.Nil(err)
	err = (client.VolumeDelete(smartvolname))
	r.Nil(err)

}

func testSmartVolumeDistribute(t *testing.T) {
	r := require.New(t)
	smartvolname := formatVolName(t.Name())

	// Too small brick size as a result of asked distribute size
	createReq := api.VolCreateReq{
		Name:               smartvolname,
		Size:               20 * gutils.MiB,
		DistributeCount:    3,
		SubvolZonesOverlap: true,
	}
	volinfo, err := client.VolumeCreate(createReq)
	r.NotNil(err)

	// create Distribute 3 Volume
	createReq = api.VolCreateReq{
		Name:            smartvolname,
		Size:            60 * gutils.MiB,
		DistributeCount: 3,
	}
	volinfo, err = client.VolumeCreate(createReq)
	r.Nil(err)
	r.Len(volinfo.Subvols, 3)
	r.Equal("Distribute", volinfo.Type.String())
	r.Len(volinfo.Subvols[0].Bricks, 1)
	r.Len(volinfo.Subvols[1].Bricks, 1)
	r.Len(volinfo.Subvols[2].Bricks, 1)

	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[0].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[1].Bricks[0].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[2].Bricks[0].Path, 16, 21))

	expandReq := api.VolExpandReq{
		Size:            60 * gutils.MiB,
		DistributeCount: 3,
	}
	_, err = client.VolumeExpand(smartvolname, expandReq)
	r.Nil(err)

	vols, err := client.Volumes(smartvolname)
	r.Nil(err)

	r.Nil(brickSizeTest(vols[0].Subvols[0].Bricks[0].Path, 37, 41))
	r.Nil(brickSizeTest(vols[0].Subvols[1].Bricks[0].Path, 37, 41))
	r.Nil(brickSizeTest(vols[0].Subvols[2].Bricks[0].Path, 37, 41))

	expandReq = api.VolExpandReq{
		Size:            210 * gutils.MiB,
		DistributeCount: 3,
	}
	_, err = client.VolumeExpand(smartvolname, expandReq)
	r.NotNil(err)

	r.Nil(client.VolumeDelete(smartvolname))
	checkZeroLvs(r)
}

func testSmartVolumeReplicate2(t *testing.T) {
	r := require.New(t)
	smartvolname := formatVolName(t.Name())
	// create Replica 2 Volume
	createReq := api.VolCreateReq{
		Name:         smartvolname,
		Size:         20 * gutils.MiB,
		ReplicaCount: 2,
	}
	volinfo, err := client.VolumeCreate(createReq)
	r.Nil(err)

	r.Len(volinfo.Subvols, 1)
	r.Equal("Replicate", volinfo.Type.String())
	r.Len(volinfo.Subvols[0].Bricks, 2)

	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[0].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[1].Path, 16, 21))

	expandReq := api.VolExpandReq{
		Size:            30 * gutils.MiB,
		DistributeCount: 1,
	}
	_, err = client.VolumeExpand(smartvolname, expandReq)
	r.Nil(err)

	vols, err := client.Volumes(smartvolname)
	r.Nil(err)

	r.Nil(brickSizeTest(vols[0].Subvols[0].Bricks[0].Path, 47, 52))
	r.Nil(brickSizeTest(vols[0].Subvols[0].Bricks[1].Path, 47, 52))

	expandReq = api.VolExpandReq{
		Size:            200 * gutils.MiB,
		DistributeCount: 1,
	}
	_, err = client.VolumeExpand(smartvolname, expandReq)
	r.NotNil(err)

	r.Nil(client.VolumeDelete(smartvolname))
	checkZeroLvs(r)
}

func testSmartVolumeReplicate3(t *testing.T) {
	r := require.New(t)

	smartvolname := formatVolName(t.Name())
	// create Replica 3 Volume
	createReq := api.VolCreateReq{
		Name:         smartvolname,
		Size:         20 * gutils.MiB,
		ReplicaCount: 3,
	}
	volinfo, err := client.VolumeCreate(createReq)
	r.Nil(err)

	r.Len(volinfo.Subvols, 1)
	r.Equal("Replicate", volinfo.Type.String())
	r.Len(volinfo.Subvols[0].Bricks, 3)
	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[0].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[1].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[2].Path, 16, 21))

	expandReq := api.VolExpandReq{
		Size:            30 * gutils.MiB,
		DistributeCount: 1,
	}
	_, err = client.VolumeExpand(smartvolname, expandReq)
	r.Nil(err)

	vols, err := client.Volumes(smartvolname)
	r.Nil(err)

	r.Nil(brickSizeTest(vols[0].Subvols[0].Bricks[0].Path, 47, 52))
	r.Nil(brickSizeTest(vols[0].Subvols[0].Bricks[1].Path, 47, 52))
	r.Nil(brickSizeTest(vols[0].Subvols[0].Bricks[1].Path, 47, 52))

	expandReq = api.VolExpandReq{
		Size:            210 * gutils.MiB,
		DistributeCount: 1,
	}
	_, err = client.VolumeExpand(smartvolname, expandReq)
	r.NotNil(err)

	r.Nil(client.VolumeDelete(smartvolname))
	checkZeroLvs(r)
}

func testSmartVolumeArbiter(t *testing.T) {
	r := require.New(t)

	smartvolname := formatVolName(t.Name())
	// create Replica 3 Arbiter Volume
	createReq := api.VolCreateReq{
		Name:         smartvolname,
		Size:         20 * gutils.MiB,
		ReplicaCount: 2,
		ArbiterCount: 1,
	}
	volinfo, err := client.VolumeCreate(createReq)
	r.Nil(err)

	r.Len(volinfo.Subvols, 1)
	r.Equal("Replicate", volinfo.Type.String())
	r.Len(volinfo.Subvols[0].Bricks, 3)
	r.Equal("Arbiter", volinfo.Subvols[0].Bricks[2].Type.String())

	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[0].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[1].Path, 16, 21))

	// TODO: Change this after arbiter calculation fix
	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[2].Path, 16, 21))

	r.Nil(client.VolumeDelete(smartvolname))
	checkZeroLvs(r)
}

func testSmartVolumeDisperse(t *testing.T) {
	r := require.New(t)

	smartvolname := formatVolName(t.Name())

	// Too small brick size as a result of asked disperse size
	createReq := api.VolCreateReq{
		Name:               smartvolname,
		Size:               20 * gutils.MiB,
		DisperseCount:      3,
		SubvolZonesOverlap: true,
	}
	volinfo, err := client.VolumeCreate(createReq)
	r.NotNil(err)

	// create Disperse Volume
	createReq = api.VolCreateReq{
		Name:          smartvolname,
		Size:          40 * gutils.MiB,
		DisperseCount: 3,
	}
	volinfo, err = client.VolumeCreate(createReq)
	r.Nil(err)

	r.Len(volinfo.Subvols, 1)
	r.Equal("Disperse", volinfo.Type.String())
	r.Len(volinfo.Subvols[0].Bricks, 3)

	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[0].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[1].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[2].Path, 16, 21))

	expandReq := api.VolExpandReq{
		Size:            30 * gutils.MiB,
		DistributeCount: 1,
	}
	_, err = client.VolumeExpand(smartvolname, expandReq)
	r.Nil(err)

	vols, err := client.Volumes(smartvolname)
	r.Nil(err)

	r.Nil(brickSizeTest(vols[0].Subvols[0].Bricks[0].Path, 27, 32))
	r.Nil(brickSizeTest(vols[0].Subvols[0].Bricks[1].Path, 27, 32))
	r.Nil(brickSizeTest(vols[0].Subvols[0].Bricks[1].Path, 27, 32))

	expandReq = api.VolExpandReq{
		Size:            240 * gutils.MiB,
		DistributeCount: 1,
	}
	_, err = client.VolumeExpand(smartvolname, expandReq)
	r.NotNil(err)

	r.Nil(client.VolumeDelete(smartvolname))
	checkZeroLvs(r)
}

func testSmartVolumeDistributeReplicate(t *testing.T) {
	r := require.New(t)

	smartvolname := formatVolName(t.Name())

	// create Distribute Replicate(2x3) Volume
	createReq := api.VolCreateReq{
		Name:               smartvolname,
		Size:               40 * gutils.MiB,
		DistributeCount:    2,
		ReplicaCount:       3,
		SubvolZonesOverlap: true,
	}
	volinfo, err := client.VolumeCreate(createReq)
	r.Nil(err)

	r.Len(volinfo.Subvols, 2)
	r.Equal("Distributed-Replicate", volinfo.Type.String())
	r.Len(volinfo.Subvols[0].Bricks, 3)
	r.Len(volinfo.Subvols[1].Bricks, 3)

	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[0].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[1].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[2].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[1].Bricks[0].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[1].Bricks[1].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[1].Bricks[2].Path, 16, 21))

	r.Nil(client.VolumeDelete(smartvolname))
	checkZeroLvs(r)
}

func testSmartVolumeDistributeDisperse(t *testing.T) {
	r := require.New(t)

	smartvolname := formatVolName(t.Name())

	// create Distribute Disperse(2x3) Volume
	createReq := api.VolCreateReq{
		Name:               smartvolname,
		Size:               80 * gutils.MiB,
		DistributeCount:    2,
		DisperseCount:      3,
		SubvolZonesOverlap: true,
	}
	volinfo, err := client.VolumeCreate(createReq)
	r.Nil(err)

	r.Len(volinfo.Subvols, 2)
	r.Equal("Distributed-Disperse", volinfo.Type.String())
	r.Len(volinfo.Subvols[0].Bricks, 3)
	r.Len(volinfo.Subvols[1].Bricks, 3)

	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[0].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[1].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[2].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[1].Bricks[0].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[1].Bricks[1].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[1].Bricks[2].Path, 16, 21))

	r.Nil(client.VolumeDelete(smartvolname))
	checkZeroLvs(r)
}

func testSmartVolumeAutoDistributeReplicate(t *testing.T) {
	r := require.New(t)

	smartvolname := formatVolName(t.Name())

	// Too small value for max-brick-size
	createReq := api.VolCreateReq{
		Name:               smartvolname,
		Size:               40 * gutils.MiB,
		ReplicaCount:       3,
		MaxBrickSize:       10 * gutils.MiB,
		SubvolZonesOverlap: true,
	}
	volinfo, err := client.VolumeCreate(createReq)
	r.NotNil(err)

	createReq = api.VolCreateReq{
		Name:               smartvolname,
		Size:               40 * gutils.MiB,
		ReplicaCount:       3,
		MaxBrickSize:       20 * gutils.MiB,
		SubvolZonesOverlap: true,
	}
	volinfo, err = client.VolumeCreate(createReq)
	r.Nil(err)

	r.Len(volinfo.Subvols, 2)
	r.Equal("Distributed-Replicate", volinfo.Type.String())
	r.Len(volinfo.Subvols[0].Bricks, 3)
	r.Len(volinfo.Subvols[1].Bricks, 3)

	r.Nil(client.VolumeDelete(smartvolname))
	checkZeroLvs(r)

	// Max-brick-size is more than request size
	createReq = api.VolCreateReq{
		Name:               smartvolname,
		Size:               20 * gutils.MiB,
		ReplicaCount:       3,
		MaxBrickSize:       30 * gutils.MiB,
		SubvolZonesOverlap: true,
	}
	volinfo, err = client.VolumeCreate(createReq)
	r.Nil(err)

	r.Len(volinfo.Subvols, 1)
	r.Equal("Replicate", volinfo.Type.String())
	r.Len(volinfo.Subvols[0].Bricks, 3)

	r.Nil(client.VolumeDelete(smartvolname))
	checkZeroLvs(r)
}

func testSmartVolumeAutoDistributeDisperse(t *testing.T) {
	r := require.New(t)

	smartvolname := formatVolName(t.Name())

	createReq := api.VolCreateReq{
		Name:               smartvolname,
		Size:               80 * gutils.MiB,
		DisperseCount:      3,
		MaxBrickSize:       20 * gutils.MiB,
		SubvolZonesOverlap: true,
	}
	volinfo, err := client.VolumeCreate(createReq)
	r.Nil(err)

	r.Len(volinfo.Subvols, 2)
	r.Equal("Distributed-Disperse", volinfo.Type.String())
	r.Len(volinfo.Subvols[0].Bricks, 3)
	r.Len(volinfo.Subvols[1].Bricks, 3)

	r.Nil(client.VolumeDelete(smartvolname))
	checkZeroLvs(r)
}

func testSmartVolumeWhenCloneExists(t *testing.T) {
	r := require.New(t)

	smartvolname := "svol1"

	createReq := api.VolCreateReq{
		Name:         smartvolname,
		Size:         20 * gutils.MiB,
		ReplicaCount: 3,
	}
	volinfo, err := client.VolumeCreate(createReq)
	r.Nil(err)

	r.Len(volinfo.Subvols, 1)
	r.Equal("Replicate", volinfo.Type.String())
	r.Len(volinfo.Subvols[0].Bricks, 3)

	err = client.VolumeStart(smartvolname, false)
	r.Nil(err)

	// Snapshot Create, Activate, Clone and delete Snapshot
	snapshotCreateReq := api.SnapCreateReq{
		VolName:  smartvolname,
		SnapName: smartvolname + "-s1",
	}
	_, err = client.SnapshotCreate(snapshotCreateReq)
	r.Nil(err, "snapshot create failed")

	var snapshotActivateReq api.SnapActivateReq

	err = client.SnapshotActivate(snapshotActivateReq, smartvolname+"-s1")
	r.Nil(err)

	snapshotCloneReq := api.SnapCloneReq{
		CloneName: smartvolname + "-c1",
	}
	_, err = client.SnapshotClone(smartvolname+"-s1", snapshotCloneReq)
	r.Nil(err, "snapshot clone failed")

	err = client.SnapshotDelete(smartvolname + "-s1")
	r.Nil(err)

	// Check number of Lvs
	nlv, err := numberOfLvs("gluster-dev-gluster_loop1")
	r.Nil(err)
	// Thinpool + brick + Clone volume's brick
	r.Equal(3, nlv)

	r.Nil(client.VolumeStop(smartvolname))

	r.Nil(client.VolumeDelete(smartvolname))

	nlv, err = numberOfLvs("gluster-dev-gluster_loop1")
	r.Nil(err)
	// Thinpool + brick + Clone volume's brick
	r.Equal(2, nlv)

	// Delete Clone Volume
	r.Nil(client.VolumeDelete(smartvolname + "-c1"))

	checkZeroLvs(r)
}

func editDevice(t *testing.T) {
	r := require.New(t)
	peerList, err := client.Peers()
	r.Nil(err)

	var deviceList []deviceapi.Info
	var peerID string
	for _, peer := range peerList {
		deviceList, err = client.DeviceList(peer.ID.String(), "")
		r.Nil(err)
		if len(deviceList) > 0 {
			peerID = peer.ID.String()
			break
		}
	}

	device := deviceList[0]
	if device.State == "enabled" {
		err = client.DeviceEdit(peerID, device.Device, "disabled")
		r.Nil(err)
	} else if device.State == "disabled" {
		err = client.DeviceEdit(peerID, device.Device, "enabled")
		r.Nil(err)
	}
	newDeviceList, err := client.DeviceList(peerID, "")
	r.Nil(err)
	for _, newDevice := range newDeviceList {
		if newDevice.Device == device.Device {
			r.NotEqual(newDevice.State, device.State)
		}
	}

	for _, peer := range peerList {
		deviceList, err := client.DeviceList(peer.ID.String(), "")
		r.Nil(err)
		for _, device := range deviceList {
			if device.State == "enabled" {
				err = client.DeviceEdit(peer.ID.String(), device.Device, "disabled")
				r.Nil(err)
			}
		}
	}
	smartvolname := formatVolName(t.Name())

	// create Distribute Replicate(2x3) Volume
	createReq := api.VolCreateReq{
		Name:               smartvolname,
		Size:               40 * gutils.MiB,
		DistributeCount:    2,
		ReplicaCount:       3,
		SubvolZonesOverlap: true,
	}
	_, err = client.VolumeCreate(createReq)
	r.NotNil(err)

	for _, peer := range peerList {
		deviceList, err := client.DeviceList(peer.ID.String(), "")
		r.Nil(err)
		for _, device := range deviceList {
			if device.State == "disabled" {
				err = client.DeviceEdit(peer.ID.String(), device.Device, "enabled")
				r.Nil(err)
			}
		}
	}

	_, err = client.VolumeCreate(createReq)
	r.Nil(err)

	r.Nil(client.VolumeDelete(smartvolname))
}

// TestSmartVolume creates a volume and starts it, runs further tests on it and
// finally deletes the volume
func TestSmartVolume(t *testing.T) {
	var err error

	r := require.New(t)

	tc, err := setupCluster(t, "./config/1.toml", "./config/2.toml", "./config/3.toml")
	r.Nil(err)
	defer teardownCluster(tc)

	client, err = initRestclient(tc.gds[0])
	r.Nil(err)
	r.NotNil(client)

	devicesDir := testTempDir(t, "devices")

	// Device Setup
	// Around 150MB will be reserved during pv/vg creation, create device with more size
	r.Nil(prepareLoopDevice(devicesDir+"/gluster_dev1.img", "1", "250M"))
	r.Nil(prepareLoopDevice(devicesDir+"/gluster_dev2.img", "2", "250M"))
	r.Nil(prepareLoopDevice(devicesDir+"/gluster_dev3.img", "3", "250M"))

	_, err = client.DeviceAdd(tc.gds[0].PeerID(), "/dev/gluster_loop1")
	r.Nil(err)
	dev, err := client.DeviceList(tc.gds[0].PeerID(), "/dev/gluster_loop1")
	r.Nil(err)
	r.Equal(dev[0].Device, "/dev/gluster_loop1")

	_, err = client.DeviceAdd(tc.gds[1].PeerID(), "/dev/gluster_loop2")
	r.Nil(err)
	dev, err = client.DeviceList(tc.gds[1].PeerID(), "/dev/gluster_loop2")
	r.Nil(err)
	r.Equal(dev[0].Device, "/dev/gluster_loop2")

	_, err = client.DeviceAdd(tc.gds[2].PeerID(), "/dev/gluster_loop3")
	r.Nil(err)
	dev, err = client.DeviceList(tc.gds[2].PeerID(), "/dev/gluster_loop3")
	r.Nil(err)
	r.Equal(dev[0].Device, "/dev/gluster_loop3")

	t.Run("Smartvol Distributed Volume", testSmartVolumeDistribute)
	t.Run("Smartvol Replicate 2 Volume", testSmartVolumeReplicate2)
	t.Run("Smartvol Replicate 3 Volume", testSmartVolumeReplicate3)
	t.Run("Smartvol Arbiter Volume", testSmartVolumeArbiter)
	t.Run("Smartvol Disperse Volume", testSmartVolumeDisperse)
	t.Run("Smartvol Distributed-Replicate Volume", testSmartVolumeDistributeReplicate)
	t.Run("Smartvol Distributed-Disperse Volume", testSmartVolumeDistributeDisperse)
	t.Run("Smartvol Auto Distributed-Replicate Volume", testSmartVolumeAutoDistributeReplicate)
	t.Run("Smartvol Auto Distributed-Disperse Volume", testSmartVolumeAutoDistributeDisperse)
	// Test dependent lvs in thinpool cases
	t.Run("Smartvol delete when clone exists", testSmartVolumeWhenCloneExists)
	t.Run("Replace Brick", testReplaceBrick)
	t.Run("Edit device", editDevice)

	// // Device Cleanup
	r.Nil(loopDevicesCleanup(t))
}
