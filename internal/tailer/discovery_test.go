package tailer

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiscoverFilesMatchesPatternInDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "app.log"), "one\n")
	mustWriteFile(t, filepath.Join(dir, "worker.log"), "two\n")
	mustWriteFile(t, filepath.Join(dir, "notes.txt"), "skip\n")

	got, err := DiscoverFiles(dir, "*.log")
	require.NoError(t, err)

	want := []string{
		filepath.Join(dir, "app.log"),
		filepath.Join(dir, "worker.log"),
	}

	require.True(t, slices.Equal(got, want), "DiscoverFiles() = %v, want %v", got, want)
}

func TestInitialOffsetStartsAtEndUnlessFromStartRequested(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	const content = "line one\nline two\n"
	mustWriteFile(t, path, content)

	offset, err := InitialOffset(path, false)
	require.NoError(t, err)
	require.Equal(t, int64(len(content)), offset)

	offset, err = InitialOffset(path, true)
	require.NoError(t, err)
	require.Zero(t, offset)
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()

	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}
