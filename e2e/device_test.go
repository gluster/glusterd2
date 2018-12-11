package e2e

import (
	"testing"

	"github.com/gluster/glusterd2/pkg/api"
	gutils "github.com/gluster/glusterd2/pkg/utils"
	deviceapi "github.com/gluster/glusterd2/plugins/device/api"

	"github.com/stretchr/testify/require"
)

func editDevice(t *testing.T) {
	r := require.New(t)
	peerList, err := client.Peers()
	r.Nil(err)

	var deviceList []deviceapi.Info
	var peerID string
	for _, peer := range peerList {
		deviceList, err = client.DeviceList(peer.ID.String(), "")
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

func testDeviceDelete(t *testing.T) {
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

	err = client.DeviceDelete(peerID, deviceList[0].Device)
	r.Nil(err)

	newDeviceList, err := client.DeviceList(peerID, "")
	r.Nil(err)

	r.Equal(len(deviceList)-1, len(newDeviceList))
}

// TestDevice creates devices in the test environment
// finally deletes the devices
func TestDevice(t *testing.T) {
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

	t.Run("Edit device", editDevice)
	t.Run("Delete device", testDeviceDelete)

	// // Device Cleanup
	r.Nil(loopDevicesCleanup(t))
}
