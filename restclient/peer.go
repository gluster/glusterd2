package restclient

import (
	"fmt"
	"strings"
)

func (c *RESTClient) PeerProbe(host string) error {
	reqBody := strings.NewReader(fmt.Sprintf(`{"addresses": ["%s"]}`, host))
	return httpRESTAction("POST", c.baseURL+"/v1/peers", reqBody, 201)
}

func (c *RESTClient) PeerDetach(host string) error {
	delURL := fmt.Sprintf(c.baseURL+"/v1/peers/%s", host)
	return httpRESTAction("DELETE", delURL, nil, 204)
}
