package restclient

import (
	"fmt"
	"net/http"
	"strings"

	deviceapi "github.com/gluster/glusterd2/plugins/device/api"
)

// DeviceAdd registers device
func (c *Client) DeviceAdd(peerid, device string) (deviceapi.AddDeviceResp, error) {
	var peerinfo deviceapi.AddDeviceResp
	req := deviceapi.AddDeviceReq{
		Device: device,
	}
	err := c.post("/v1/devices/"+peerid, req, http.StatusOK, &peerinfo)
	return peerinfo, err
}

// DeviceList lists the devices
func (c *Client) DeviceList(peerid string) ([]deviceapi.Info, error) {
	var deviceList deviceapi.ListDeviceResp
	url := fmt.Sprintf("/v1/devices/%s", peerid)
	err := c.get(url, nil, http.StatusOK, &deviceList)
	return deviceList, err
}

// DeviceEdit edits device
func (c *Client) DeviceEdit(peerid, device, state string) error {
	req := deviceapi.EditDeviceReq{
		State: state,
	}
	device = strings.TrimLeft(device, "/")
	url := fmt.Sprintf("/v1/devices/%s/%s", peerid, device)
	err := c.post(url, req, http.StatusOK, nil)
	return err
}
