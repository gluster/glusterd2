package e2e

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"testing"

	"github.com/gluster/glusterd2/pkg/api"

	"github.com/stretchr/testify/require"
)

var (
	testvols = []string{"testvol1", "testvol2"}

	pidcount  = make(map[int]int)
	portcount = make(map[int]int)
)

func TestBrickMuxVolumeOps(t *testing.T) {
	var err error

	r := require.New(t)

	gds, err = setupCluster("./config/1.toml")
	r.Nil(err)
	defer teardownCluster(gds)

	client = initRestclient(gds[0].ClientAddress)

	var brickMuxOpts = map[string]string{
		"cluster.brick-multiplex":        "on",
		"cluster.max-bricks-per-process": "4",
	}

	err = client.GlobalOptionSet(api.GlobalOptionReq{
		Options: brickMuxOpts,
	})
	r.Nil(err)

	tmpDir, err = ioutil.TempDir("", t.Name())
	r.Nil(err)
	t.Logf("Using temp dir: %s", tmpDir)
	//defer os.RemoveAll(tmpDir)

	// Create the volumes
	r.Nil(testVolsCreate(testvols), "volume creation failed")

	for _, volname := range testvols {
		err = client.VolumeStart(volname)
		r.Nil(err, "volume %s start failed", volname)
	}

	r.Nil(testVolStatusAndUpdateCounts(testvols))

	r.True(len(pidcount) == 2, fmt.Sprintf("Pid count: %d", len(pidcount)))
	r.True(len(portcount) == 2, fmt.Sprintf("Port count: %d", len(portcount)))

	r.Nil(testVolumesMounts(testvols))

	err = client.GlobalOptionSet(api.GlobalOptionReq{
		Options: map[string]string{"cluster.max-bricks-per-process": "5"},
	})
	r.Nil(err)

	r.Nil(testVolsCreate([]string{"testvol3"}), "volume creation failed")

	err = client.VolumeStart("testvol3")
	r.Nil(err, "volume testvol3 start failed")

	r.Nil(testVolStatusAndUpdateCounts(append(testvols, "testvol3")))

	r.True(len(pidcount) == 2, fmt.Sprintf("Pid count: %d", len(pidcount)))
	r.True(len(portcount) == 2, fmt.Sprintf("Port count: %d", len(portcount)))

	r.Nil(testVolumesMounts(append(testvols, "testvol3")))

	err = client.VolumeStop("testvol2")
	r.Nil(err, "volume testvol2 stop failed")

	err = client.VolumeStart("testvol2")
	r.Nil(err, "volume testvol2 start failed")

	r.Nil(testVolStatusAndUpdateCounts(append(testvols, "testvol3")))
	r.True(len(pidcount) == 2, fmt.Sprintf("Pid count: %d", len(pidcount)))
	r.True(len(portcount) == 2, fmt.Sprintf("Port count: %d", len(portcount)))

}

// testVolumesMounts checks if volumes mount successfully and unmounts them
func testVolumesMounts(testvols []string) error {
	mntPath, err := ioutil.TempDir(tmpDir, "mnt")
	if err != nil {
		return err
	}
	defer os.RemoveAll(mntPath)

	for _, volname := range testvols {
		host, _, _ := net.SplitHostPort(gds[0].ClientAddress)
		mntCmd := exec.Command("mount", "-t", "glusterfs", host+":"+volname, mntPath)
		umntCmd := exec.Command("umount", mntPath)

		err = mntCmd.Run()
		if err != nil {
			return err
		}

		err = umntCmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func testVolStatusAndUpdateCounts(testvols []string) error {
	for _, volname := range testvols {
		volstatus, err := client.BricksStatus(volname)
		if err != nil {
			return err
		}

		for _, brickstatus := range volstatus {
			pid := brickstatus.Pid
			port := brickstatus.Port
			if _, found := pidcount[pid]; found {
				pidcount[pid] = pidcount[pid] + 1
			} else {
				pidcount[pid] = 1
			}

			if _, found := portcount[port]; found {
				portcount[port] = portcount[port] + 1
			} else {
				portcount[port] = 1
			}
		}
	}
	return nil
}

func testVolsCreate(testvols []string) error {
	for _, volname := range testvols {
		var brickPaths []string
		for i := 1; i <= 3; i++ {
			brickPath, err := ioutil.TempDir(tmpDir, "brick")
			if err != nil {
				return err
			}
			brickPaths = append(brickPaths, brickPath)
		}

		// create a 1x3 replicate volume
		createReq := api.VolCreateReq{
			Name: volname,
			Subvols: []api.SubvolReq{
				{
					ReplicaCount: 3,
					Type:         "replicate",
					Bricks: []api.BrickReq{
						{PeerID: gds[0].PeerID(), Path: brickPaths[0]},
						{PeerID: gds[0].PeerID(), Path: brickPaths[1]},
						{PeerID: gds[0].PeerID(), Path: brickPaths[2]},
					},
				},
			},
			Force: true,
		}
		_, err := client.VolumeCreate(createReq)
		if err != nil {
			return err
		}
	}

	return nil
}
