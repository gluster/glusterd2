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
	"github.com/gluster/glusterd2/version"

	"github.com/dgrijalva/jwt-go"
)

const (
	expireSeconds = 120
	clientTimeout = 30
)

// Client represents Glusterd2 REST Client
type Client struct {
	baseURL  string
	username string
	password string
	cacert   string
	insecure bool

	// Add to the identifier to further specify the client
	// using the api.
	agent      string
	originArgs []string
}

// New creates new instance of Glusterd REST Client
func New(baseURL string, username string, password string, cacert string, insecure bool) *Client {
	return &Client{baseURL, username, password, cacert, insecure,
		fmt.Sprintf("GlusterD2-rest-client/%v", version.GlusterdVersion),
		[]string{}}
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
	c.setAgent(req)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Close = true

	// Set Authorization if username and password is not empty string
	if c.username != "" && c.password != "" {
		req.Header.Set("Authorization", "bearer "+getAuthToken(c.username, c.password))
	}

	c.sendOriginArgs(req)

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
		Timeout:   clientTimeout * time.Second}

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

func (c *Client) setAgent(req *http.Request) {
	req.Header.Set("User-Agent",
		fmt.Sprintf("%v (Go-http-client/1.1)", c.agent))
}

// ExtendAgent adds the given string to the client's user agent
// by prefixing it to the existing agent information.
// This allows client programs to identify themselves more than
// just something using the rest client api.
func (c *Client) ExtendAgent(a string) {
	// new additions to the agent identifier go to the
	// beginning of the string. It is meant to read like:
	// foo (based on) bar (based on) baz, etc...
	c.agent = fmt.Sprintf("%v %v", a, c.agent)
}

func (c *Client) sendOriginArgs(req *http.Request) {
	if len(c.originArgs) != 0 {
		req.Header.Set("X-Gluster-Origin-Args",
			fmt.Sprintf("1:%#v", strings.Join(c.originArgs, " ")))
	}
}

// SetOriginArgs provides a way for tools using this library to
// inform the server what arguments were provided to generate
// the api call(s). The contents of the array are meant only for
// human interpretation but will generally be command line args.
func (c *Client) SetOriginArgs(a []string) {
	c.originArgs = a
}
