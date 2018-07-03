package gdctx

import (
	"os"
	"testing"

	config "github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestSetHostnameAndIP(t *testing.T) {
	err := SetHostnameAndIP()
	assert.Nil(t, err)
}

func TestGenerateLocalAuthToken(t *testing.T) {
	err := GenerateLocalAuthToken()
	assert.Nil(t, err)

	config.Set("restauth", true)
	config.Set("localstatedir", "/tmp/gd2test/")
	err = GenerateLocalAuthToken()
	assert.NotNil(t, err)

	config.Set("localstatedir", "")
	err = GenerateLocalAuthToken()
	assert.Nil(t, err)
	os.Remove("auth")
}
