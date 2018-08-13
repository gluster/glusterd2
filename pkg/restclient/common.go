package restclient

import (
	"bytes"
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

// ClientFunc receives a Client and overrides its members
type ClientFunc func(*Client) error

// WithTLSConfig overrides http.Client member with a client created
// using specified TLS configuration.
func WithTLSConfig(tlsOpts *TLSOptions) ClientFunc {
	return func(client *Client) error {
		tlsConfig, err := NewTLSConfig(tlsOpts)
		if err != nil {
			return err
		}
		tr := &http.Transport{
			DisableCompression: true,
			DisableKeepAlives:  true,
			TLSClientConfig:    tlsConfig,
		}
		httpClient := &http.Client{
			Transport: tr,
		}
		client.httpClient = httpClient
		return nil
	}
}

// WithHTTPClient overrides http Client with specified one.
func WithHTTPClient(httpClient *http.Client) ClientFunc {
	return func(client *Client) error {
		if httpClient != nil {
			client.httpClient = httpClient
		}
		return nil
	}
}

// WithBaseURL overrides Client base url with specified one
func WithBaseURL(url string) ClientFunc {
	return func(client *Client) error {
		client.baseURL = url
		return nil
	}
}

// WithUsername overrides Client username with specified one
func WithUsername(username string) ClientFunc {
	return func(client *Client) error {
		client.username = username
		return nil
	}
}

// WithPassword overrides Client password with specified one
func WithPassword(password string) ClientFunc {
	return func(client *Client) error {
		client.password = password
		return nil
	}
}

// WithTimeOut overrides Client timeout with specified one
func WithTimeOut(timeout time.Duration) ClientFunc {
	return func(client *Client) error {
		client.httpClient.Timeout = timeout
		return nil
	}
}

// Client represents Glusterd2 REST Client
type Client struct {
	baseURL     string
	username    string
	password    string
	timeout     time.Duration
	httpClient  *http.Client
	lastRespErr *http.Response
}

// NewClientWithOpts initializes a default Glusterd2 REST Client.
// It takes functors to modify it while creating.
// For e.g., `NewClientWithOpts(WithBaseURL(...),WithUsername(...))`
// We can also initialize custom http Client using WithHTTPClient(...)
// to send request.
func NewClientWithOpts(opts ...ClientFunc) (*Client, error) {
	client := &Client{
		httpClient: http.DefaultClient,
	}
	for _, fn := range opts {
		if err := fn(client); err != nil {
			return client, err
		}
	}
	return client, nil
}

// New creates new instance of Glusterd2 REST Client
// Deprecated : Use NewClientWithOpts(...)
func New(baseURL, username, password, cacert string, insecure bool) (*Client, error) {
	return NewClientWithOpts(
		WithBaseURL(baseURL),
		WithTLSConfig(&TLSOptions{CaCertFile: cacert, InsecureSkipVerify: insecure}),
		WithUsername(username),
		WithPassword(password),
		WithTimeOut(defaultClientTimeout*time.Second),
	)
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
// Deprecated : use `WithTimeOut(...)`
func (c *Client) SetTimeout(timeout time.Duration) {
	WithTimeOut(timeout)(c)
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
	req, err := c.buildRequest(method, url, input)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
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

func (c *Client) buildRequest(method string, url string, input interface{}) (*http.Request, error) {
	url = fmt.Sprintf("%s%s", c.baseURL, url)
	var body io.Reader
	if input != nil {
		reqBody, err := json.Marshal(input)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(reqBody)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Close = true

	// Set Authorization if username and password is not empty string
	if c.username != "" && c.password != "" {
		c.setAuthToken(req)
	}
	return req, nil
}

//Ping checks glusterd2 service status
func (c *Client) Ping() error {
	return c.get("/ping", nil, http.StatusOK, nil)
}
