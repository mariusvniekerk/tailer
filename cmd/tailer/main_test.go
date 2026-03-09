package main

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/mvanniekerk/tailer/internal/tailer"
	"github.com/stretchr/testify/require"
)

func TestNewRootCmdPassesFlagsToRunner(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	called := false
	cmd := newRootCmd(func(ctx context.Context, cfg tailer.Config, out io.Writer) error {
		called = true

		require.Equal(t, dir, cfg.Dir)
		require.Equal(t, "*.txt", cfg.Pattern)
		require.Equal(t, 250*time.Millisecond, cfg.PollInterval)
		require.True(t, cfg.FromStart)
		require.NotNil(t, out)

		return nil
	}, io.Discard, io.Discard)

	cmd.SetArgs([]string{
		"--dir", dir,
		"--pattern", "*.txt",
		"--poll-interval", "250ms",
		"--from-start",
	})

	require.NoError(t, cmd.Execute())
	require.True(t, called)
}
