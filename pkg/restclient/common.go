package restclient

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gluster/glusterd2/pkg/api"
)

// Client represents Glusterd2 REST Client
type Client struct {
	baseURL  string
	username string
	password string
}

// New creates new instance of Glusterd REST Client
func New(baseURL string, username string, password string) *Client {
	return &Client{baseURL, username, password}
}

func parseHTTPError(jsonData []byte) string {
	var errstr api.HttpError
	err := json.Unmarshal(jsonData, &errstr)
	if err != nil {
		return ""
	}
	return errstr.Error
}

func (c *Client) do(method string, url string, respType string, data interface{}, expectStatusCode int) (string, error) {
	url = fmt.Sprintf("%s%s", c.baseURL, url)

	var body io.Reader
	if data != nil {
		reqBody, marshalErr := json.Marshal(data)
		if marshalErr != nil {
			return "", marshalErr
		}
		body = strings.NewReader(string(reqBody))
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return "", err
	}
	if respType == "xml" || respType == "json" {
		req.Header.Set("Accept", "application/"+respType)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err1 := http.DefaultClient.Do(req)
	if err1 != nil {
		return "", err1
	}

	defer resp.Body.Close()
	output, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		return "", err2
	}
	if resp.StatusCode != expectStatusCode {
		return "", &UnexpectedStatusError{"Unexpected Status", expectStatusCode, resp.StatusCode, parseHTTPError(output)}
	}
	return parseHTTPError(output), nil
}

func (c *Client) action(method string, url string, data interface{}, expectStatusCode int) error {
	_, err := c.do(method, url, "", data, expectStatusCode)
	return err
}

func (c *Client) getJSON(url string) (string, error) {
	return c.do("GET", url, "json", nil, http.StatusOK)
}

func (c *Client) getXML(url string) (string, error) {
	return c.do("GET", url, "xml", nil, http.StatusOK)
}
