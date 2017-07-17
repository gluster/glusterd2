package restclient

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const baseURL = "http://localhost:24007"

type RESTClient struct {
	baseURL  string
	username string
	password string
}

func NewRESTClient(baseURL string, username string, password string) *RESTClient {
	return &RESTClient{baseURL, username, password}
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

	if resp.StatusCode == 404 {
		return "", NewAPIUnsupportedError()
	}
	defer resp.Body.Close()
	output, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		return "", err2
	}
	if resp.StatusCode != expectStatusCode {
		return "", NewAPIUnexpectedStatusCode(expectStatusCode, resp.StatusCode, string(output))
	}
	return string(output), nil
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
