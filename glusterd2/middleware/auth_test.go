package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"io/ioutil"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gluster/glusterd2/glusterd2/gdctx"
	config "github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetAuthSecret(t *testing.T) {
	secret := getAuthSecret("test")
	assert.Empty(t, secret)

	config.Set("restauth", true)
	config.Set("localstatedir", "")
	err := gdctx.GenerateLocalAuthToken()
	assert.Nil(t, err)
	os.Remove("auth")

	secret = getAuthSecret("glustercli")
	assert.NotNil(t, secret)
}

func getAuthToken(username string, password string, r *http.Request) {
	// Generate qsh
	qshstring := "GET&/"
	hash := sha256.New()
	hash.Write([]byte(qshstring))
	// Create Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": username,
		"exp": time.Now().Add(time.Second * time.Duration(120)).Unix(),
		"qsh": hex.EncodeToString(hash.Sum(nil)),
	})
	// Sign the token
	signedtoken, err := token.SignedString([]byte(password))
	if err != nil {
		return
	}
	r.Header.Set("Authorization", "bearer "+signedtoken)
}

func generateLocalauthtoken() {

	config.Set("restauth", true)
	config.Set("localstatedir", "")
	gdctx.GenerateLocalAuthToken()

}

func TestAuth(t *testing.T) {

	ts := httptest.NewServer(Auth(GetTestHandler()))
	gdctx.RESTAPIAuthEnabled = false
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	gdctx.RESTAPIAuthEnabled = true
	resp, err = http.Get(ts.URL)
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusUnauthorized)

	generateLocalauthtoken()
	secret, err := ioutil.ReadFile("auth")
	assert.Nil(t, err)

	client := http.Client{}
	req, err := http.NewRequest("GET", ts.URL, nil)
	assert.Nil(t, err)

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	getAuthToken("glustercli", string(secret), req)
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	getAuthToken("testuser", string(secret), req)
	resp, err = client.Do(req)
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusUnauthorized)

	os.Remove("auth")
}

func GetTestHandler() http.HandlerFunc {
	fn := func(rw http.ResponseWriter, req *http.Request) {

	}
	return http.HandlerFunc(fn)
}
