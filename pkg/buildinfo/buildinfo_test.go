package buildinfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	result := String()

	assert.Contains(t, result, "GophKeeper")
	assert.Contains(t, result, Version)
	assert.Contains(t, result, BuildDate)
}

func TestString_DefaultValues(t *testing.T) {
	result := String()

	assert.Equal(t, "GophKeeper dev (built unknown)", result)
}

func TestString_CustomValues(t *testing.T) {
	originalVersion := Version
	originalBuildDate := BuildDate
	defer func() {
		Version = originalVersion
		BuildDate = originalBuildDate
	}()

	Version = "1.0.0"
	BuildDate = "2024-01-15"

	result := String()

	assert.Equal(t, "GophKeeper 1.0.0 (built 2024-01-15)", result)
}
