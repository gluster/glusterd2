package restclient

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClientWithOpts(t *testing.T) {
	assert := assert.New(t)
	httpClient := &http.Client{}
	client, err := NewClientWithOpts(
		WithHTTPClient(httpClient),
		WithTimeOut(time.Second*10),
		WithBaseURL("http://localhost:8080"),
		WithUsername("testuser"),
		WithPassword("testpasswod"),
	)
	assert.Nil(err, "creating GD2 client")
	assert.Equal(client.baseURL, "http://localhost:8080")
	assert.Equal(client.password, "testpasswod")
	assert.Equal(client.username, "testuser")
	assert.Equal(time.Second*10, client.httpClient.Timeout)
	assert.Nil(client.LastErrorResponse())
}

func TestNewClientWithOpts_Error(t *testing.T) {
	assert := assert.New(t)
	httpClient := &http.Client{}
	client, err := NewClientWithOpts(
		WithHTTPClient(httpClient),
		WithTLSConfig(&TLSOptions{InsecureSkipVerify: false, CaCertFile: "wrongFile.pem"}),
	)
	assert.NotNil(err)
	assert.NotEqual(client.username, "testuser")
}
