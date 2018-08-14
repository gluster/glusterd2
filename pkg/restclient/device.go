package restclient

import (
	"net/http"

	deviceapi "github.com/gluster/glusterd2/plugins/device/api"
)

// DeviceAdd registers devices
func (c *Client) DeviceAdd(peerid string, device string) (deviceapi.AddDeviceResp, *http.Response, error) {
	var peerinfo deviceapi.AddDeviceResp
	req := deviceapi.AddDeviceReq{
		Device: device,
	}
	resp, err := c.post("/v1/devices/"+peerid, req, http.StatusOK, &peerinfo)
	return peerinfo, resp, err
}
