package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteStatedump(t *testing.T) {
	r := require.New(t)

	dir, err := ioutil.TempDir("", t.Name())
	r.Nil(err)
	defer os.RemoveAll(dir)

	WriteStatedump(dir)

	filePattern := fmt.Sprintf("glusterd2.%s.dump.*", strconv.Itoa(os.Getpid()))
	matches, err := filepath.Glob(filepath.Join(dir, filePattern))
	r.Nil(err)
	r.NotEmpty(matches)
}
