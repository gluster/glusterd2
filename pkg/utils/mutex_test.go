package utils

import "testing"
import "github.com/stretchr/testify/assert"

func TestTryLock(t *testing.T) {
	lock := MutexWithTry{}
	defer lock.Unlock()

	l := lock.TryLock()
	assert.True(t, l)

	l = lock.TryLock()
	assert.False(t, l)

}
