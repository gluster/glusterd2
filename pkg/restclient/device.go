package restclient

import (
	"fmt"
	"net/http"
	"strings"

	deviceapi "github.com/gluster/glusterd2/plugins/device/api"
)

// DeviceAdd registers device
func (c *Client) DeviceAdd(peerid, device string) (deviceapi.AddDeviceResp, error) {
	var deviceinfo deviceapi.AddDeviceResp
	req := deviceapi.AddDeviceReq{
		Device: device,
	}
	err := c.post("/v1/devices/"+peerid, req, http.StatusCreated, &deviceinfo)
	return deviceinfo, err
}

// DeviceList lists the devices
func (c *Client) DeviceList(peerid, device string) ([]deviceapi.Info, error) {
	var deviceList deviceapi.ListDeviceResp
	url := "/v1/devices"
	if peerid != "" {
		url = fmt.Sprintf("%s/%s", url, peerid)
		if device != "" {
			url = fmt.Sprintf("%s/%s", url, strings.TrimPrefix(device, "/"))
		}
	}

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
