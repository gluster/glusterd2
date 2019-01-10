package e2e

import (
	"os"
	"testing"

	"github.com/gluster/glusterd2/pkg/api"
	gutils "github.com/gluster/glusterd2/pkg/utils"
	deviceapi "github.com/gluster/glusterd2/plugins/device/api"

	"github.com/stretchr/testify/require"
)

// Replace brick test
func testReplaceBrickLoop(t *testing.T) {
	r := require.New(t)
	smartvolname := formatVolName(t.Name())
	// create Distribute 3 Volume
	createReq := api.VolCreateReq{
		Name:            smartvolname,
		Size:            60 * gutils.MiB,
		DistributeCount: 2,
		ProvisionerType: api.ProvisionerTypeLoop,
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

func testSmartVolumeDistributeLoop(t *testing.T) {
	r := require.New(t)
	smartvolname := formatVolName(t.Name())

	// Too small brick size as a result of asked distribute size
	createReq := api.VolCreateReq{
		Name:               smartvolname,
		Size:               20 * gutils.MiB,
		DistributeCount:    3,
		SubvolZonesOverlap: true,
		ProvisionerType:    api.ProvisionerTypeLoop,
	}
	volinfo, err := client.VolumeCreate(createReq)
	r.NotNil(err)

	// create Distribute 3 Volume
	createReq = api.VolCreateReq{
		Name:            smartvolname,
		Size:            60 * gutils.MiB,
		DistributeCount: 3,
		ProvisionerType: api.ProvisionerTypeLoop,
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

	r.Nil(client.VolumeDelete(smartvolname))
}

func testSmartVolumeReplicate2Loop(t *testing.T) {
	r := require.New(t)
	smartvolname := formatVolName(t.Name())
	// create Replica 2 Volume
	createReq := api.VolCreateReq{
		Name:            smartvolname,
		Size:            20 * gutils.MiB,
		ReplicaCount:    2,
		ProvisionerType: api.ProvisionerTypeLoop,
	}
	volinfo, err := client.VolumeCreate(createReq)
	r.Nil(err)

	r.Len(volinfo.Subvols, 1)
	r.Equal("Replicate", volinfo.Type.String())
	r.Len(volinfo.Subvols[0].Bricks, 2)

	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[0].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[1].Path, 16, 21))

	r.Nil(client.VolumeDelete(smartvolname))
}

func testSmartVolumeReplicate3Loop(t *testing.T) {
	r := require.New(t)

	smartvolname := formatVolName(t.Name())
	// create Replica 3 Volume
	createReq := api.VolCreateReq{
		Name:            smartvolname,
		Size:            20 * gutils.MiB,
		ReplicaCount:    3,
		ProvisionerType: api.ProvisionerTypeLoop,
	}
	volinfo, err := client.VolumeCreate(createReq)
	r.Nil(err)

	r.Len(volinfo.Subvols, 1)
	r.Equal("Replicate", volinfo.Type.String())
	r.Len(volinfo.Subvols[0].Bricks, 3)
	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[0].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[1].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[2].Path, 16, 21))

	r.Nil(client.VolumeDelete(smartvolname))
}

func testSmartVolumeArbiterLoop(t *testing.T) {
	r := require.New(t)

	smartvolname := formatVolName(t.Name())
	// create Replica 3 Arbiter Volume
	createReq := api.VolCreateReq{
		Name:            smartvolname,
		Size:            20 * gutils.MiB,
		ReplicaCount:    2,
		ArbiterCount:    1,
		ProvisionerType: api.ProvisionerTypeLoop,
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
}

func testSmartVolumeDisperseLoop(t *testing.T) {
	r := require.New(t)

	smartvolname := formatVolName(t.Name())

	// Too small brick size as a result of asked disperse size
	createReq := api.VolCreateReq{
		Name:               smartvolname,
		Size:               20 * gutils.MiB,
		DisperseCount:      3,
		SubvolZonesOverlap: true,
		ProvisionerType:    api.ProvisionerTypeLoop,
	}
	volinfo, err := client.VolumeCreate(createReq)
	r.NotNil(err)

	// create Disperse Volume
	createReq = api.VolCreateReq{
		Name:            smartvolname,
		Size:            40 * gutils.MiB,
		DisperseCount:   3,
		ProvisionerType: api.ProvisionerTypeLoop,
	}
	volinfo, err = client.VolumeCreate(createReq)
	r.Nil(err)

	r.Len(volinfo.Subvols, 1)
	r.Equal("Disperse", volinfo.Type.String())
	r.Len(volinfo.Subvols[0].Bricks, 3)

	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[0].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[1].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[2].Path, 16, 21))

	r.Nil(client.VolumeDelete(smartvolname))
}

func testSmartVolumeDistributeReplicateLoop(t *testing.T) {
	r := require.New(t)

	smartvolname := formatVolName(t.Name())

	// create Distribute Replicate(2x3) Volume
	createReq := api.VolCreateReq{
		Name:               smartvolname,
		Size:               40 * gutils.MiB,
		DistributeCount:    2,
		ReplicaCount:       3,
		SubvolZonesOverlap: true,
		ProvisionerType:    api.ProvisionerTypeLoop,
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
}

func testSmartVolumeDistributeDisperseLoop(t *testing.T) {
	r := require.New(t)

	smartvolname := formatVolName(t.Name())

	// create Distribute Disperse(2x3) Volume
	createReq := api.VolCreateReq{
		Name:               smartvolname,
		Size:               80 * gutils.MiB,
		DistributeCount:    2,
		DisperseCount:      3,
		SubvolZonesOverlap: true,
		ProvisionerType:    api.ProvisionerTypeLoop,
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
}

func testSmartVolumeAutoDistributeReplicateLoop(t *testing.T) {
	r := require.New(t)

	smartvolname := formatVolName(t.Name())

	// Too small value for max-brick-size
	createReq := api.VolCreateReq{
		Name:               smartvolname,
		Size:               40 * gutils.MiB,
		ReplicaCount:       3,
		MaxBrickSize:       10 * gutils.MiB,
		SubvolZonesOverlap: true,
		ProvisionerType:    api.ProvisionerTypeLoop,
	}
	volinfo, err := client.VolumeCreate(createReq)
	r.NotNil(err)

	createReq = api.VolCreateReq{
		Name:               smartvolname,
		Size:               40 * gutils.MiB,
		ReplicaCount:       3,
		MaxBrickSize:       20 * gutils.MiB,
		SubvolZonesOverlap: true,
		ProvisionerType:    api.ProvisionerTypeLoop,
	}
	volinfo, err = client.VolumeCreate(createReq)
	r.Nil(err)

	r.Len(volinfo.Subvols, 2)
	r.Equal("Distributed-Replicate", volinfo.Type.String())
	r.Len(volinfo.Subvols[0].Bricks, 3)
	r.Len(volinfo.Subvols[1].Bricks, 3)

	r.Nil(client.VolumeDelete(smartvolname))

	// Max-brick-size is more than request size
	createReq = api.VolCreateReq{
		Name:               smartvolname,
		Size:               20 * gutils.MiB,
		ReplicaCount:       3,
		MaxBrickSize:       30 * gutils.MiB,
		SubvolZonesOverlap: true,
		ProvisionerType:    api.ProvisionerTypeLoop,
	}
	volinfo, err = client.VolumeCreate(createReq)
	r.Nil(err)

	r.Len(volinfo.Subvols, 1)
	r.Equal("Replicate", volinfo.Type.String())
	r.Len(volinfo.Subvols[0].Bricks, 3)

	r.Nil(client.VolumeDelete(smartvolname))
}

func testSmartVolumeAutoDistributeDisperseLoop(t *testing.T) {
	r := require.New(t)

	smartvolname := formatVolName(t.Name())

	createReq := api.VolCreateReq{
		Name:               smartvolname,
		Size:               80 * gutils.MiB,
		DisperseCount:      3,
		MaxBrickSize:       20 * gutils.MiB,
		SubvolZonesOverlap: true,
		ProvisionerType:    api.ProvisionerTypeLoop,
	}
	volinfo, err := client.VolumeCreate(createReq)
	r.Nil(err)

	r.Len(volinfo.Subvols, 2)
	r.Equal("Distributed-Disperse", volinfo.Type.String())
	r.Len(volinfo.Subvols[0].Bricks, 3)
	r.Len(volinfo.Subvols[1].Bricks, 3)

	r.Nil(client.VolumeDelete(smartvolname))
}

func editDeviceLoop(t *testing.T) {
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
		ProvisionerType:    api.ProvisionerTypeLoop,
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

// TestSmartVolumeLoop creates a volume and starts it, runs further tests on it and
// finally deletes the volume
func TestSmartVolumeLoop(t *testing.T) {
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

	r.Nil(os.MkdirAll(devicesDir+"/exports1", 0700))
	r.Nil(os.MkdirAll(devicesDir+"/exports2", 0700))
	r.Nil(os.MkdirAll(devicesDir+"/exports3", 0700))

	_, err = client.DeviceAdd(tc.gds[0].PeerID(), devicesDir+"/exports1", api.ProvisionerTypeLoop)
	r.Nil(err)
	dev, err := client.DeviceList(tc.gds[0].PeerID(), devicesDir+"/exports1")
	r.Nil(err)
	r.Equal(dev[0].Device, devicesDir+"/exports1")

	_, err = client.DeviceAdd(tc.gds[1].PeerID(), devicesDir+"/exports2", api.ProvisionerTypeLoop)
	r.Nil(err)
	dev, err = client.DeviceList(tc.gds[1].PeerID(), devicesDir+"/exports2")
	r.Nil(err)
	r.Equal(dev[0].Device, devicesDir+"/exports2")

	_, err = client.DeviceAdd(tc.gds[2].PeerID(), devicesDir+"/exports3", api.ProvisionerTypeLoop)
	r.Nil(err)
	dev, err = client.DeviceList(tc.gds[2].PeerID(), devicesDir+"/exports3")
	r.Nil(err)
	r.Equal(dev[0].Device, devicesDir+"/exports3")

	t.Run("Smartvol Distributed Volume Loop", testSmartVolumeDistributeLoop)
	t.Run("Smartvol Replicate 2 Volume Loop", testSmartVolumeReplicate2Loop)
	t.Run("Smartvol Replicate 3 Volume Loop", testSmartVolumeReplicate3Loop)
	t.Run("Smartvol Arbiter Volume Loop", testSmartVolumeArbiterLoop)
	t.Run("Smartvol Disperse Volume Loop", testSmartVolumeDisperseLoop)
	t.Run("Smartvol Distributed-Replicate Volume Loop", testSmartVolumeDistributeReplicateLoop)
	t.Run("Smartvol Distributed-Disperse Volume Loop", testSmartVolumeDistributeDisperseLoop)
	t.Run("Smartvol Auto Distributed-Replicate Volume Loop", testSmartVolumeAutoDistributeReplicateLoop)
	t.Run("Smartvol Auto Distributed-Disperse Volume Loop", testSmartVolumeAutoDistributeDisperseLoop)
	t.Run("Replace Brick Loop", testReplaceBrickLoop)
	t.Run("Edit device Loop", editDeviceLoop)

	// // Device Cleanup
	cleanupAllBrickMounts(t)
}
