package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mariusvniekerk/tailer/internal/tailer"
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

func TestModulePathMatchesPublishedRepository(t *testing.T) {
	t.Parallel()

	goModPath := filepath.Join("..", "..", "go.mod")
	goModBytes, err := os.ReadFile(goModPath)
	require.NoError(t, err)

	require.True(
		t,
		strings.HasPrefix(string(goModBytes), "module github.com/mariusvniekerk/tailer\n"),
		"go.mod should declare the published GitHub module path",
	)
}
