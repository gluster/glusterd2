package e2e

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/gluster/glusterd2/pkg/api"

	"github.com/stretchr/testify/require"
)

// TestQuota creates a volume and starts it, runs further quota enable on it
// and finally deletes the volume
func TestQuota(t *testing.T) {
	var err error
	var brickPaths []string
	r := require.New(t)

	gds, err := setupCluster("./config/1.toml", "./config/2.toml")
	r.Nil(err)
	defer teardownCluster(gds)

	brickDir, err := ioutil.TempDir(baseLocalStateDir, t.Name())
	r.Nil(err)
	defer os.RemoveAll(brickDir)
	t.Logf("Using temp dir: %s", brickDir)

	volumeName := formatVolName(t.Name())

	for i := 1; i <= 4; i++ {
		brickPath, err := ioutil.TempDir(brickDir, "brick")
		r.Nil(err)
		brickPaths = append(brickPaths, brickPath)
	}

	client := initRestclient(gds[0])

	// create 2x2 dist-rep volume
	createReq := api.VolCreateReq{
		Name: volumeName,
		Subvols: []api.SubvolReq{
			{
				ReplicaCount: 2,
				Type:         "replicate",
				Bricks: []api.BrickReq{
					{PeerID: gds[0].PeerID(), Path: brickPaths[0]},
					{PeerID: gds[1].PeerID(), Path: brickPaths[1]},
				},
			},
			{
				Type:         "replicate",
				ReplicaCount: 2,
				Bricks: []api.BrickReq{
					{PeerID: gds[0].PeerID(), Path: brickPaths[2]},
					{PeerID: gds[1].PeerID(), Path: brickPaths[3]},
				},
			},
		},
		Force: true,
	}

	_, err = client.VolumeCreate(createReq)
	r.Nil(err)

	// test Quota on dist-rep volume
	t.Run("Quota-enable", func(t *testing.T) { testQuotaEnable(t, gds) })

	r.Nil(client.VolumeDelete(volumeName))
}

func testQuotaEnable(t *testing.T, gds []*gdProcess) {
	var err error
	r := require.New(t)

	// form the pidfile path
	pidpath := path.Join(gds[0].Rundir, "quotad.pid")

	quotaKey := "quota.enable"
	var optionReqOff api.VolOptionReq
	optionReqOff.Advanced = true

	optionReqOff.Options = map[string]string{quotaKey: "off"}

	// Quota not enabled: no quotad should be there
	err = client.VolumeSet(volname, optionReqOff)
	r.Contains(err.Error(), "quotad is not enabled")

	// Checking if the quotad is not running
	r.False(isProcessRunning(pidpath))

	var optionReqOn api.VolOptionReq
	optionReqOn.Advanced = true

	// Enable quota
	quotaKey = "quota.enable"
	optionReqOn.Options = map[string]string{quotaKey: "on"}
	// Quotad should be there
	r.Nil(client.VolumeSet(volname, optionReqOn))

	// Checking if the quotad is running
	r.True(isProcessRunning(pidpath))

	// check the error for enabling it again
	err = client.VolumeSet(volname, optionReqOn)
	r.Contains(err.Error(), "process is already running")

	// Checking if the quotad is running
	r.True(isProcessRunning(pidpath))

	// Disable quota
	r.Nil(client.VolumeSet(volname, optionReqOff))

	// Checking if the quotad is not running
	r.False(isProcessRunning(pidpath))

	// Check the error for disabling it again.
	err = client.VolumeSet(volname, optionReqOff)
	r.Contains(err.Error(), "quotad is not enabled")

	// Checking if the quotad is not running
	r.False(isProcessRunning(pidpath))

	// Stop Volume
	r.Nil(client.VolumeStop(volname), "Volume stop failed")
	// delete volume
	err = client.VolumeDelete(volname)
	r.Nil(err)

}
