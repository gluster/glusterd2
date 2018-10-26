package e2e

import (
	"fmt"
	"syscall"
	"testing"

	"github.com/gluster/glusterd2/pkg/api"

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
	for i := 1; i < 3; i++ {
		nlv, err := numberOfLvs(fmt.Sprintf("vg-dev-gluster_loop%d", i))
		r.Nil(err)
		if err == nil {
			r.Equal(0, nlv)
		}
	}
}

func testSmartVolumeDistribute(t *testing.T) {
	r := require.New(t)
	smartvolname := formatVolName(t.Name())
	// create Distribute 3 Volume
	createReq := api.VolCreateReq{
		Name:            smartvolname,
		Size:            60,
		DistributeCount: 3,
	}
	volinfo, err := client.VolumeCreate(createReq)
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
	checkZeroLvs(r)
}

func testSmartVolumeReplicate2(t *testing.T) {
	r := require.New(t)
	smartvolname := formatVolName(t.Name())
	// create Replica 2 Volume
	createReq := api.VolCreateReq{
		Name:         smartvolname,
		Size:         20,
		ReplicaCount: 2,
	}
	volinfo, err := client.VolumeCreate(createReq)
	r.Nil(err)

	r.Len(volinfo.Subvols, 1)
	r.Equal("Replicate", volinfo.Type.String())
	r.Len(volinfo.Subvols[0].Bricks, 2)

	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[0].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[1].Path, 16, 21))

	r.Nil(client.VolumeDelete(smartvolname))
	checkZeroLvs(r)
}

func testSmartVolumeReplicate3(t *testing.T) {
	r := require.New(t)

	smartvolname := formatVolName(t.Name())
	// create Replica 3 Volume
	createReq := api.VolCreateReq{
		Name:         smartvolname,
		Size:         20,
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

	r.Nil(client.VolumeDelete(smartvolname))
	checkZeroLvs(r)
}

func testSmartVolumeArbiter(t *testing.T) {
	r := require.New(t)

	smartvolname := formatVolName(t.Name())
	// create Replica 3 Arbiter Volume
	createReq := api.VolCreateReq{
		Name:         smartvolname,
		Size:         20,
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

	// create Disperse Volume
	createReq := api.VolCreateReq{
		Name:          smartvolname,
		Size:          40,
		DisperseCount: 3,
	}
	volinfo, err := client.VolumeCreate(createReq)
	r.Nil(err)

	r.Len(volinfo.Subvols, 1)
	r.Equal("Disperse", volinfo.Type.String())
	r.Len(volinfo.Subvols[0].Bricks, 3)

	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[0].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[1].Path, 16, 21))
	r.Nil(brickSizeTest(volinfo.Subvols[0].Bricks[2].Path, 16, 21))

	r.Nil(client.VolumeDelete(smartvolname))
	checkZeroLvs(r)
}

func testSmartVolumeDistributeReplicate(t *testing.T) {
	r := require.New(t)

	smartvolname := formatVolName(t.Name())

	// create Distribute Replicate(2x3) Volume
	createReq := api.VolCreateReq{
		Name:               smartvolname,
		Size:               40,
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
		Size:               80,
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

	_, err = client.DeviceAdd(tc.gds[1].PeerID(), "/dev/gluster_loop2")
	r.Nil(err)

	_, err = client.DeviceAdd(tc.gds[2].PeerID(), "/dev/gluster_loop3")
	r.Nil(err)

	t.Run("Smartvol Distributed Volume", testSmartVolumeDistribute)
	t.Run("Smartvol Replicate 2 Volume", testSmartVolumeReplicate2)
	t.Run("Smartvol Replicate 3 Volume", testSmartVolumeReplicate3)
	t.Run("Smartvol Arbiter Volume", testSmartVolumeArbiter)
	t.Run("Smartvol Disperse Volume", testSmartVolumeDisperse)
	t.Run("Smartvol Distributed-Replicate Volume", testSmartVolumeDistributeReplicate)
	t.Run("Smartvol Distributed-Disperse Volume", testSmartVolumeDistributeDisperse)

	// // Device Cleanup
	r.Nil(loopDevicesCleanup(t))
}
