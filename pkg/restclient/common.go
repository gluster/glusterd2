package restclient

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gluster/glusterd2/pkg/utils"

	"github.com/dgrijalva/jwt-go"
)

const (
	expireSeconds        = 120
	defaultClientTimeout = 30 // in seconds
)

// Client represents Glusterd2 REST Client
type Client struct {
	baseURL     string
	username    string
	password    string
	cacert      string
	insecure    bool
	timeout     time.Duration
	lastRespErr *http.Response
}

// LastErrorResponse returns the last error response received by this
// client from glusterd2. Please note that the Body of the response has
// been read and drained.
func (c *Client) LastErrorResponse() *http.Response {
	return c.lastRespErr
}

// SetTimeout sets the overall client timeout which includes the time taken
// from setting up TCP connection till client finishes reading the response
// body.
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// New creates new instance of Glusterd REST Client
func New(baseURL, username, password, cacert string, insecure bool) *Client {
	return &Client{
		baseURL:  baseURL,
		username: username,
		password: password,
		cacert:   cacert,
		insecure: insecure,
		timeout:  defaultClientTimeout * time.Second,
	}
}

func (c *Client) setAuthToken(r *http.Request) {
	// Create Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		// Set issuer
		"iss": c.username,
		// Set expiration
		"exp": time.Now().Add(time.Second * expireSeconds).Unix(),
		// Set qsh
		"qsh": utils.GenerateQsh(r),
	})
	// Sign the token
	signedtoken, err := token.SignedString([]byte(c.password))
	if err != nil {
		return
	}
	r.Header.Set("Authorization", "bearer "+signedtoken)
	return
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

func (c *Client) do(method string, url string, input interface{}, expectStatusCode int, output interface{}) error {
	url = fmt.Sprintf("%s%s", c.baseURL, url)

	var body io.Reader
	if input != nil {
		reqBody, marshalErr := json.Marshal(input)
		if marshalErr != nil {
			return marshalErr
		}
		body = bytes.NewReader(reqBody)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Close = true

	// Set Authorization if username and password is not empty string
	if c.username != "" && c.password != "" {
		c.setAuthToken(req)
	}

	tr := &http.Transport{
		DisableCompression: true,
		DisableKeepAlives:  true,
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

	client := &http.Client{
		Transport: tr,
		Timeout:   c.timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectStatusCode {
		// FIXME: We should may be rather look for 4xx or 5xx series
		// to determine that we got an error response instead of
		// comparing to what's expected ?
		c.lastRespErr = resp
		return newHTTPErrorResponse(resp)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// If a response struct is specified, unmarshall the json response
	// body into the response struct provided.
	if output != nil {
		return json.Unmarshal(b, output)
	}

	return nil
}

//Ping checks glusterd2 service status
func (c *Client) Ping() error {
	return c.get("/ping", nil, http.StatusOK, nil)
}
