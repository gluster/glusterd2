package utils

import (
	"testing"

	"github.com/gluster/glusterd2/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestGetTypeString(t *testing.T) {
	resp := GetTypeString((*api.PeerGetResp)(nil))
	assert.Equal(t, resp, "api.PeerGetResp")

}
