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

func (c *Client) post(url string, data interface{}, expectStatusCode int, output interface{}) error {
	return c.do("POST", url, data, expectStatusCode, output)
}

func (c *Client) put(url string, data interface{}, expectStatusCode int, output interface{}) error {
	return c.do("PUT", url, data, expectStatusCode, output)
}

func (c *Client) get(url string, data interface{}, expectStatusCode int, output interface{}) error {
	return c.do("GET", url, data, expectStatusCode, output)
}

func (c *Client) del(url string, data interface{}, expectStatusCode int, output interface{}) error {
	return c.do("DELETE", url, data, expectStatusCode, output)
}

func (c *Client) do(method string, url string, data interface{}, expectStatusCode int, output interface{}) error {
	url = fmt.Sprintf("%s%s", c.baseURL, url)

	var body io.Reader
	if data != nil {
		reqBody, marshalErr := json.Marshal(data)
		if marshalErr != nil {
			return marshalErr
		}
		body = strings.NewReader(string(reqBody))
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	resp, err1 := http.DefaultClient.Do(req)
	if err1 != nil {
		return err1
	}

	defer resp.Body.Close()
	outputRaw, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		return err2
	}
	if resp.StatusCode != expectStatusCode {
		return &UnexpectedStatusError{"Unexpected Status", expectStatusCode, resp.StatusCode, parseHTTPError(outputRaw)}
	}

	if output != nil {
		return json.Unmarshal(outputRaw, output)
	}

	return nil
}
