package utils

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	config "github.com/spf13/viper"

	"github.com/stretchr/testify/assert"
)

func TestWriteStatedump(t *testing.T) {
	filename := fmt.Sprintf("glusterd2.%s.dump.%s",
		strconv.Itoa(os.Getpid()), strconv.Itoa(int(time.Now().Unix())))
	WriteStatedump()
	assert.FileExists(t, filename)
	os.Remove(filename)

	config.Set("rundir", "/abc")
	WriteStatedump()
}
