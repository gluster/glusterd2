package e2e

import (
	"fmt"
	"os"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/gluster/glusterd2/pkg/api"

	"github.com/stretchr/testify/require"
)

// TestBrickMux tests brick multiplexing.
func TestBrickMux(t *testing.T) {
	var err error

	r := require.New(t)

	tc, err := setupCluster(t, "./config/1.toml")
	r.Nil(err)
	defer teardownCluster(tc)

	client, err = initRestclient(tc.gds[0])
	r.Nil(err)
	r.NotNil(client)

	// Turn on brick mux cluster option
	optReq := api.ClusterOptionReq{
		Options: map[string]string{"cluster.brick-multiplex": "on"},
	}
	err = client.ClusterOptionSet(optReq)
	r.Nil(err)

	// Create a 1 x 3 volume
	var brickPaths []string
	for i := 1; i <= 3; i++ {
		brickPath := testTempDir(t, "brick")
		brickPaths = append(brickPaths, brickPath)
	}

	volname1 := formatVolName(t.Name())
	volname2 := volname1 + strconv.Itoa(2)

	createReq := api.VolCreateReq{
		Name: volname1,
		Subvols: []api.SubvolReq{
			{
				Type: "distribute",
				Bricks: []api.BrickReq{
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[0]},
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[1]},
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[2]},
				},
			},
		},
		Force: true,
	}
	_, err = client.VolumeCreate(createReq)
	r.Nil(err)

	// start the volume
	err = client.VolumeStart(volname1, false)
	r.Nil(err)

	// check bricks status and confirm that bricks have been multiplexed
	bstatus, err := client.BricksStatus(volname1)
	r.Nil(err)

	// NOTE: Track these variables through-out the test.
	pid := bstatus[0].Pid
	port := bstatus[0].Port

	for _, b := range bstatus {
		r.Equal(pid, b.Pid)
		r.Equal(port, b.Port)
	}

	// create another compatible volume

	for i := 4; i <= 5; i++ {
		brickPath := testTempDir(t, "brick")
		brickPaths = append(brickPaths, brickPath)
	}

	createReq = api.VolCreateReq{
		Name: volname2,
		Subvols: []api.SubvolReq{
			{
				Type: "distribute",
				Bricks: []api.BrickReq{
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[3]},
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[4]},
				},
			},
		},
		Force: true,
	}
	_, err = client.VolumeCreate(createReq)
	r.Nil(err)

	err = client.VolumeStart(volname2, false)
	r.Nil(err)

	// check bricks status and confirm that bricks have been multiplexed
	// onto bricks of the first volume
	bstatus, err = client.BricksStatus(volname2)
	r.Nil(err)

	// the pid and port variables now point to values from the old original volume
	for _, b := range bstatus {
		r.Equal(pid, b.Pid)
		r.Equal(port, b.Port)
	}

	// kill the brick from first volume into which all the brick have been multiplexed
	process, err := os.FindProcess(pid)
	r.Nil(err, fmt.Sprintf("failed to find bricks pid: %s", err))
	err = process.Signal(syscall.Signal(15))
	r.Nil(err, fmt.Sprintf("failed to kill bricks: %s", err))

	time.Sleep(time.Second * 1)
	bstatus, err = client.BricksStatus(volname1)
	r.Nil(err)
	r.Equal(bstatus[0].Pid, 0)
	r.Equal(bstatus[0].Port, 0)

	// Second volume's bricks should become offline since brick from first volume has been killed
	bstatus, err = client.BricksStatus(volname2)
	r.Nil(err)
	for _, b := range bstatus {
		r.Equal(b.Online, false)
	}

	// force start the first and second volume
	err = client.VolumeStart(volname1, true)
	r.Nil(err)

	err = client.VolumeStart(volname2, true)
	r.Nil(err)

	// first brick from first volume should now all bricks of first volume
	// should be  be multiplexed into a new pid
	bstatus, err = client.BricksStatus(volname1)
	r.Nil(err)

	pid = bstatus[0].Pid
	port = bstatus[0].Port

	// force start the second volume and the bricks of second volume should
	// now be multiplexed into the pid in which bricks of first volume are multiplexed
	bstatus, err = client.BricksStatus(volname2)
	r.Nil(err)

	for _, b := range bstatus {
		r.Equal(pid, b.Pid)
		r.Equal(port, b.Port)
	}

	// stop the second volume, make it incompatible for multiplexing and start it again.
	// this should start the bricks as separate processes.
	r.Nil(client.VolumeStop(volname2))

	voloptReq := api.VolOptionReq{
		Options: map[string]string{"io-stats.count-fop-hits": "on"},
	}
	voloptReq.AllowAdvanced = true
	err = client.VolumeSet(volname2, voloptReq)
	r.Nil(err)

	err = client.VolumeStart(volname2, false)
	r.Nil(err)

	bstatus, err = client.BricksStatus(volname2)
	r.Nil(err)

	// the pid and port variables point to values from the old values
	// the bricks should have different values for pid and port as they
	// are no longer multiplexed
	for _, b := range bstatus {
		r.NotEqual(pid, b.Pid)
		r.NotEqual(port, b.Port)
	}

	r.Nil(client.VolumeStop(volname2))
	r.Nil(client.VolumeStop(volname1))

	r.Nil(client.VolumeDelete(volname2))
	r.Nil(client.VolumeDelete(volname1))

	for i := 6; i <= 36; i++ {
		brickPath := testTempDir(t, "brick")
		brickPaths = append(brickPaths, brickPath)
	}

	// Create 10 volumes and start all 10
	// making all brick multiplexed into first brick of
	// first volume.
	index := 5
	for i := 1; i <= 10; i++ {
		createReq := api.VolCreateReq{
			Name: volname1 + strconv.Itoa(i),
			Subvols: []api.SubvolReq{
				{
					Type: "distribute",
					Bricks: []api.BrickReq{
						{PeerID: tc.gds[0].PeerID(), Path: brickPaths[index]},
						{PeerID: tc.gds[0].PeerID(), Path: brickPaths[index+1]},
						{PeerID: tc.gds[0].PeerID(), Path: brickPaths[index+2]},
					},
				},
			},
			Force: true,
		}
		_, err = client.VolumeCreate(createReq)
		r.Nil(err)

		// start the volume
		err = client.VolumeStart(volname1+strconv.Itoa(i), false)
		r.Nil(err)

		index = index + 3
	}

	// Check if the multiplexing was successful
	for i := 1; i <= 10; i++ {

		bstatus, err = client.BricksStatus(volname1 + strconv.Itoa(i))
		r.Nil(err)

		if i == 1 {
			pid = bstatus[0].Pid
			port = bstatus[0].Port
		} else {
			for _, b := range bstatus {
				r.Equal(pid, b.Pid)
				r.Equal(port, b.Port)
			}
		}

	}

	// Stop glusterd2 instance and kill the glusterfsd into
	// which all bricks were multiplexed
	r.Nil(tc.gds[0].Stop())
	process, err = os.FindProcess(pid)
	r.Nil(err, fmt.Sprintf("failed to find brick pid: %s", err))
	err = process.Signal(syscall.Signal(15))
	r.Nil(err, fmt.Sprintf("failed to kill brick: %s", err))

	// Spawn glusterd2 instance
	gd, err := spawnGlusterd(t, "./config/1.toml", false)
	r.Nil(err)
	r.True(gd.IsRunning())

	// Check if all the bricks are multiplexed into the first brick
	// of first volume, this time with a different pid and port.
	for i := 1; i <= 10; i++ {

		bstatus, err = client.BricksStatus(volname1 + strconv.Itoa(i))
		r.Nil(err)
		if i == 1 {
			pid = bstatus[0].Pid
			port = bstatus[0].Port
		}
		for _, b := range bstatus {
			r.Equal(pid, b.Pid)
			r.Equal(port, b.Port)
		}
	}

	for i := 1; i <= 10; i++ {
		r.Nil(client.VolumeStop(volname1 + strconv.Itoa(i)))
		r.Nil(client.VolumeDelete(volname1 + strconv.Itoa(i)))
	}

	// Turn on brick mux max-bricks-per-process cluster option
	optReq = api.ClusterOptionReq{
		Options: map[string]string{"cluster.max-bricks-per-process": "5"},
	}
	err = client.ClusterOptionSet(optReq)
	r.Nil(err)

	for i := 36; i <= 200; i++ {
		brickPath := testTempDir(t, "brick")
		brickPaths = append(brickPaths, brickPath)
	}

	/* Test  for Testing max-bricks-per-process constraint while
	multiplexing*/

	// Create 10 volumes and start all 10
	// making all brick multiplexed with the constraint that
	// max-bricks-per-process is 5

	index = 37
	for i := 1; i <= 10; i++ {

		createReq := api.VolCreateReq{
			Name: volname1 + strconv.Itoa(i),
			Subvols: []api.SubvolReq{
				{
					Type: "distribute",
					Bricks: []api.BrickReq{
						{PeerID: tc.gds[0].PeerID(), Path: brickPaths[index]},
						{PeerID: tc.gds[0].PeerID(), Path: brickPaths[index+1]},
					},
				},
			},
			Force: true,
		}

		if i%2 != 0 {
			createReq = api.VolCreateReq{
				Name: volname1 + strconv.Itoa(i),
				Subvols: []api.SubvolReq{
					{
						ReplicaCount: 2,
						Type:         "replicate",
						Bricks: []api.BrickReq{
							{PeerID: tc.gds[0].PeerID(), Path: brickPaths[index]},
							{PeerID: tc.gds[0].PeerID(), Path: brickPaths[index+1]},
						},
					},
				},
				Force: true,
			}

		}
		_, err = client.VolumeCreate(createReq)
		r.Nil(err)
		// start the volume
		err = client.VolumeStart(volname1+strconv.Itoa(i), false)
		r.Nil(err)

		index = index + 2
	}

	// pidMap and portMap just to mantain count of every pid and port of
	// bricks from all 5 volumes.
	pidMap := make(map[int]int)
	portMap := make(map[int]int)
	for i := 1; i <= 10; i++ {
		bstatus, err := client.BricksStatus(volname1 + strconv.Itoa(i))
		r.Nil(err)
		for _, b := range bstatus {
			if _, ok := pidMap[b.Pid]; ok {
				pidMap[b.Pid]++
			} else {
				pidMap[b.Pid] = 1
			}

			if _, ok := portMap[b.Port]; ok {
				portMap[b.Port]++
			} else {
				portMap[b.Port] = 1
			}

		}
	}

	// Check if all pid's and ports's have count = 2 as mentioned in
	// max-bricks-per-process
	for _, v := range pidMap {
		r.Equal(v, 5)
	}

	for _, v := range portMap {
		r.Equal(v, 5)
	}

	pidMap = make(map[int]int)
	portMap = make(map[int]int)
	index = 100
	for i := 1; i <= 10; i++ {
		if i%2 != 0 {
			expandReq := api.VolExpandReq{
				Bricks: []api.BrickReq{
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[index]},
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[index+1]},
				},
				Force: true,
			}
			_, err := client.VolumeExpand(volname1+strconv.Itoa(i), expandReq)
			r.Nil(err)

			index = index + 2
		} else {
			expandReq := api.VolExpandReq{
				Bricks: []api.BrickReq{
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[index]},
				},
				Force: true,
			}
			_, err := client.VolumeExpand(volname1+strconv.Itoa(i), expandReq)
			r.Nil(err)

			index = index + 1
		}
		// Wait for added bricks' glusterfsds to Sign In
		time.Sleep(5 * time.Millisecond)

		// re-populate pid and port mapping
		bstatus, err := client.BricksStatus(volname1 + strconv.Itoa(i))
		r.Nil(err)
		for _, b := range bstatus {
			if _, ok := pidMap[b.Pid]; ok {
				pidMap[b.Pid]++
			} else {
				pidMap[b.Pid] = 1
			}

			if _, ok := portMap[b.Port]; ok {
				portMap[b.Port]++
			} else {
				portMap[b.Port] = 1
			}
		}
	}

	for _, v := range pidMap {
		r.Equal(v, 5)
	}

	for _, v := range portMap {
		r.Equal(v, 5)
	}

	pidMap = make(map[int]int)
	portMap = make(map[int]int)
	for i := 1; i <= 10; i++ {
		if i%2 != 0 {
			expandReq := api.VolExpandReq{
				ReplicaCount: 3,
				Bricks: []api.BrickReq{
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[index]},
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[index+1]},
				},
				Force: true,
			}
			_, err := client.VolumeExpand(volname1+strconv.Itoa(i), expandReq)
			r.Nil(err)

			// Wait for new bricks glusterfsds to Sign In
			time.Sleep(5 * time.Millisecond)

			// re-populate pid and port mapping
			bstatus, err := client.BricksStatus(volname1 + strconv.Itoa(i))
			r.Nil(err)
			for _, b := range bstatus {
				if _, ok := pidMap[b.Pid]; ok {
					pidMap[b.Pid]++
				} else {
					pidMap[b.Pid] = 1
				}

				if _, ok := portMap[b.Port]; ok {
					portMap[b.Port]++
				} else {
					portMap[b.Port] = 1
				}

			}
			index = index + 2
		}
	}

	for _, v := range pidMap {
		r.Equal(v, 5)
	}

	for _, v := range portMap {
		r.Equal(v, 5)
	}

	r.Nil(gd.Stop())
	for k := range pidMap {
		process, err := os.FindProcess(k)
		r.Nil(err, fmt.Sprintf("failed to find brick pid: %s", err))
		err = process.Signal(syscall.Signal(15))
		r.Nil(err, fmt.Sprintf("failed to kill brick: %s", err))
	}

	// Spawn glusterd2 instance
	gd, err = spawnGlusterd(t, "./config/1.toml", false)
	r.Nil(err)
	r.True(gd.IsRunning())

	// Wait for GD2 instance and all glusterfsds to spawn up
	time.Sleep(5 * time.Second)

	pidMap = make(map[int]int)
	portMap = make(map[int]int)
	for i := 1; i <= 10; i++ {
		bstatus, err := client.BricksStatus(volname1 + strconv.Itoa(i))
		r.Nil(err)
		for _, b := range bstatus {
			if _, ok := pidMap[b.Pid]; ok {
				pidMap[b.Pid]++
			} else {
				pidMap[b.Pid] = 1
			}

			if _, ok := portMap[b.Port]; ok {
				portMap[b.Port]++
			} else {
				portMap[b.Port] = 1
			}

		}
	}

	// Check if all pid's and ports's have count = 5 as mentioned in
	// max-bricks-per-process
	for _, v := range pidMap {
		r.Equal(v, 5)
	}

	for _, v := range portMap {
		r.Equal(v, 5)
	}

	for i := 1; i <= 10; i++ {
		r.Nil(client.VolumeStop(volname1 + strconv.Itoa(i)))
		r.Nil(client.VolumeDelete(volname1 + strconv.Itoa(i)))
	}

	// Create two volumes with different options, so that bricks from these
	// two volumes are multiplexed into bricks from their own volume. Also,
	// check if among three bricks of a volume 2 bricks have same pid and
	// port while 1 brick has a different pid and port, since num  of bricks
	// are 3 and max-bricks-per-process is set as 2.

	// Turn on brick mux cluster option
	optReq = api.ClusterOptionReq{
		Options: map[string]string{"cluster.max-bricks-per-process": "2"},
	}
	err = client.ClusterOptionSet(optReq)
	r.Nil(err)

	createReq = api.VolCreateReq{
		Name: volname1,
		Subvols: []api.SubvolReq{
			{
				ReplicaCount: 3,
				Type:         "replicate",
				Bricks: []api.BrickReq{
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[51]},
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[52]},
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[53]},
				},
			},
		},
		Force: true,
	}
	_, err = client.VolumeCreate(createReq)
	r.Nil(err)

	// start the volume
	err = client.VolumeStart(volname1, false)
	r.Nil(err)

	createReq = api.VolCreateReq{
		Name: volname2,
		Subvols: []api.SubvolReq{
			{
				Type: "distribute",
				Bricks: []api.BrickReq{
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[48]},
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[49]},
					{PeerID: tc.gds[0].PeerID(), Path: brickPaths[50]},
				},
			},
		},
		Force: true,
	}
	_, err = client.VolumeCreate(createReq)
	r.Nil(err)

	// Setting an option in second volume so that  second volume doesn't
	// multiplex its brick into first volume
	var optionReq api.VolOptionReq
	optionReq.Options = map[string]string{"io-stats.count-fop-hits": "on"}
	optionReq.AllowAdvanced = true

	r.Nil(client.VolumeSet(volname2, optionReq))

	// start the volume
	err = client.VolumeStart(volname2, false)
	r.Nil(err)

	bstatus, err = client.BricksStatus(volname1)
	r.Nil(err)

	// Keep track of used unique pids and ports used in multiplexing bricks
	// of  volname1 and calculate length length of maps, which should be equal to 2
	pidMap = make(map[int]int)
	portMap = make(map[int]int)
	for _, b := range bstatus {
		pidMap[b.Pid] = 1
		portMap[b.Port] = 1
	}
	r.Equal(len(pidMap), 2)
	r.Equal(len(portMap), 2)

	bstatus2, err := client.BricksStatus(volname2)
	r.Nil(err)

	// Keep track of used unique pids and ports used in multiplexing bricks
	// of  volname1 and calculate length length of maps, which should be equal to 2
	pidMap = make(map[int]int)
	portMap = make(map[int]int)
	for _, b := range bstatus2 {
		pidMap[b.Pid] = 1
		portMap[b.Port] = 1
	}
	r.Equal(len(pidMap), 2)
	r.Equal(len(portMap), 2)

	r.Nil(client.VolumeStop(volname1))
	r.Nil(client.VolumeDelete(volname1))

	r.Nil(client.VolumeStop(volname2))
	r.Nil(client.VolumeDelete(volname2))

	r.Nil(gd.Stop())
}
