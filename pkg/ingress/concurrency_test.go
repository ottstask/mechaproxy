package ingress

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcurrency(t *testing.T) {
	l := newLimiter(2)
	ok := l.acquire()
	assert.True(t, ok)
	ok = l.acquire()
	assert.True(t, ok)

	ok = l.acquire()
	assert.True(t, !ok)

	l.release()
	ok = l.acquire()
	assert.True(t, ok)
}
