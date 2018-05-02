package volgen

import "testing"
import config "github.com/spf13/viper"

func TestSetDefaults(t *testing.T) {
	SetDefaults()
	config.Set("templatesdir", "/tmp/templates")
	SetDefaults()
}
