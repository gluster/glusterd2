package restclient

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gluster/glusterd2/pkg/api"

	"github.com/dgrijalva/jwt-go"
)

var (
	expireSeconds = 120
)

// Client represents Glusterd2 REST Client
type Client struct {
	baseURL  string
	username string
	password string
	cacert   string
	insecure bool
}

// New creates new instance of Glusterd REST Client
func New(baseURL string, username string, password string, cacert string, insecure bool) *Client {
	return &Client{baseURL, username, password, cacert, insecure}
}

func parseHTTPError(jsonData []byte) string {
	var errstr api.HTTPError
	err := json.Unmarshal(jsonData, &errstr)
	if err != nil {
		return ""
	}
	return errstr.Error
}

func getAuthToken(username string, password string) string {
	// Create the Claims
	claims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Second * time.Duration(expireSeconds)).Unix(),
		Issuer:    username,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(password))
	if err != nil {
		return ""
	}

	return ss
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

	// Set Authorization if username and password is not empty string
	if c.username != "" && c.password != "" {
		req.Header.Set("Authorization", "bearer "+getAuthToken(c.username, c.password))
	}

	tr := &http.Transport{
		DisableCompression:    true,
		DisableKeepAlives:     true,
		ResponseHeaderTimeout: 3 * time.Second,
	}

	if c.cacert != "" || c.insecure {
		caCertPool := x509.NewCertPool()
		if caCert, err := ioutil.ReadFile(c.cacert); err != nil {
			if !c.insecure {
				return err
			}
		} else {
			caCertPool.AppendCertsFromPEM(caCert)
		}
		tr.TLSClientConfig = &tls.Config{
			RootCAs:            caCertPool,
			InsecureSkipVerify: c.insecure,
		}
	}

	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	outputRaw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != expectStatusCode {
		return &UnexpectedStatusError{"Unexpected Status", expectStatusCode, resp.StatusCode, parseHTTPError(outputRaw)}
	}

	if output != nil {
		return json.Unmarshal(outputRaw, output)
	}

	return nil
}
