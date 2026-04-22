package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateReadMode_NonInteractiveWithoutTTY(t *testing.T) {
	err := validateReadMode(true, false)
	require.NoError(t, err)
}

func TestValidateReadMode_TTYRequired(t *testing.T) {
	err := validateReadMode(false, false)
	require.Error(t, err)
}

func TestValidateReadMode_InteractiveTTY(t *testing.T) {
	err := validateReadMode(false, true)
	require.NoError(t, err)
}
