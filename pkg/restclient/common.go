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
	"strings"
	"time"

	"github.com/gluster/glusterd2/pkg/api"

	"github.com/dgrijalva/jwt-go"
)

const (
	expireSeconds        = 120
	defaultClientTimeout = 30 // in seconds
)

// Client represents Glusterd2 REST Client
type Client struct {
	baseURL  string
	username string
	password string
	cacert   string
	insecure bool
	timeout  time.Duration
}

// SetTimeout sets the overall client timeout which includes the time taken
// from setting up TCP connection till client finishes reading the response
// body.
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// New creates new instance of Glusterd REST Client
func New(baseURL string, username string, password string, cacert string, insecure bool) *Client {
	return &Client{
		baseURL:  baseURL,
		username: username,
		password: password,
		cacert:   cacert,
		insecure: insecure,
		timeout:  defaultClientTimeout * time.Second,
	}
}

func parseHTTPError(jsonData []byte) string {
	var errResp api.ErrorResp
	err := json.Unmarshal(jsonData, &errResp)
	if err != nil {
		return err.Error()
	}

	var buffer bytes.Buffer
	for _, apiErr := range errResp.Errors {
		switch api.ErrorCode(apiErr.Code) {
		case api.ErrTxnStepFailed:
			buffer.WriteString(fmt.Sprintf(
				"Transaction step %s failed on peer %s with error: %s\n",
				apiErr.Fields["step"], apiErr.Fields["peer-id"], apiErr.Fields["error"]))
		default:
			buffer.WriteString(apiErr.Message)
		}
	}

	return buffer.String()
}

func getAuthToken(username string, password string) string {
	// Create the Claims
	claims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Second * expireSeconds).Unix(),
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
	req.Close = true

	// Set Authorization if username and password is not empty string
	if c.username != "" && c.password != "" {
		req.Header.Set("Authorization", "bearer "+getAuthToken(c.username, c.password))
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
		Timeout:   c.timeout}

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

//Ping checks glusterd2 service status
func (c *Client) Ping() error {
	return c.get("/ping", nil, http.StatusOK, nil)
}
