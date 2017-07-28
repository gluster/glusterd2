package restclient

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

const baseURL = "http://localhost:24007"

// RESTClient represents Glusterd2 REST Client
type RESTClient struct {
	baseURL  string
	username string
	password string
}

type httpError struct {
	Error string `json:"Error"`
}

// NewRESTClient creates new instance of RESTClient
func NewRESTClient(baseURL string, username string, password string) *RESTClient {
	return &RESTClient{baseURL, username, password}
}

func parseHTTPError(jsonData []byte) string {
	var errstr httpError
	err := json.Unmarshal(jsonData, &errstr)
	if err != nil {
		return ""
	}
	return errstr.Error
}

func httpRequest(method string, url string, respType string, body io.Reader, expectStatusCode int) (string, error) {
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
		return "", raiseAPIUnexpectedStatusCodeError(expectStatusCode, resp.StatusCode, parseHTTPError(output))
	}
	return parseHTTPError(output), nil
}

func httpRESTAction(method string, url string, body io.Reader, expectStatusCode int) error {
	_, err := httpRequest(method, url, "", body, expectStatusCode)
	return err
}

func httpGETJSON(url string) (string, error) {
	return httpRequest("GET", url, "json", nil, 200)
}

func httpGETXML(url string) (string, error) {
	return httpRequest("GET", url, "xml", nil, 200)
}
