package gdctx

import (
	"os"
	"testing"

	config "github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestInitUUID(t *testing.T) {
	err := InitUUID()
	assert.Nil(t, err)

	config.Set("localstatedir", "/tmp/gd2test/test")
	err = InitUUID()
	assert.NotNil(t, err)

	os.Remove("uuid.toml")

}

func TestSave(t *testing.T) {
	defer os.Remove("uuid.toml")
	config.Set("localstatedir", "")
	cfg := &UUIDConfig{}

	err := cfg.save()
	assert.Nil(t, err)

	config.Set("localstatedir", "/tmp/gd2test/test")
	err = cfg.save()
	assert.NotNil(t, err)
}

func TestReload(t *testing.T) {
	cfg := &UUIDConfig{}

	err := cfg.reload()
	assert.Nil(t, err)
}
