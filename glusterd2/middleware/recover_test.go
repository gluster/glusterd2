package middleware

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecover(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
	handler := Recover(h)
	assert.NotNil(t, handler)
}
