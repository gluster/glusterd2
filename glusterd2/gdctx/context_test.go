package gdctx

import (
	"context"
	"testing"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"

	log "github.com/sirupsen/logrus"
)

func TestReqID(t *testing.T) {
	var ctx context.Context
	reqID := uuid.NewRandom()

	ctx = WithReqID(ctx, reqID)
	id := GetReqID(ctx)
	assert.Equal(t, id, reqID)
}

func TestWithReqLogger(t *testing.T) {
	var ctx context.Context
	lo := log.WithField("test", "test")

	ctx = WithReqLogger(ctx, lo)
	assert.NotNil(t, ctx)

	newlog := GetReqLogger(ctx)
	assert.NotNil(t, newlog)

}
