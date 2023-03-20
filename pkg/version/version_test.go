package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVersion(t *testing.T) {
	assert.Equal(t, "v1.0.0-beta", GetVersion())

	version = "v0.1.0"
	assert.Equal(t, "v0.1.0", GetVersion())
}
