package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStartCommandOutput(t *testing.T) {
	cmd := newRootCmd("")
	b := bytes.NewBufferString("")

	cmd.SetArgs([]string{"start"})
	cmd.SetOut(b)

	cmdErr := cmd.RunE(cmd, nil)
	require.NoError(t, cmdErr)
}
