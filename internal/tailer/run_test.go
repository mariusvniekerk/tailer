package tailer

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRunEmitsAppendedAndNewFileLines(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	existingPath := filepath.Join(dir, "app.log")
	mustWriteFile(t, existingPath, "seed\n")

	cfg := Config{
		Dir:          dir,
		Pattern:      "*.log",
		PollInterval: 20 * time.Millisecond,
		FromStart:    false,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var out lockedBuffer
	errCh := make(chan error, 1)
	go func() {
		errCh <- Run(ctx, cfg, &out)
	}()

	time.Sleep(3 * cfg.PollInterval)

	appendFile(t, existingPath, "after-start\n")
	waitForContains(t, &out, existingPath+": after-start")

	newPath := filepath.Join(dir, "next.log")
	mustWriteFile(t, newPath, "created-now\n")
	waitForContains(t, &out, newPath+": created-now")

	cancel()

	select {
	case err := <-errCh:
		require.True(t, err == nil || err == context.Canceled, "Run() returned error: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not exit after cancel")
	}
}

func TestRunEmitsLinesAfterDeleteAndRecreate(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	mustWriteFile(t, path, "seed\n")

	cfg := Config{
		Dir:          dir,
		Pattern:      "*.log",
		PollInterval: 20 * time.Millisecond,
		FromStart:    false,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var out lockedBuffer
	errCh := make(chan error, 1)
	go func() {
		errCh <- Run(ctx, cfg, &out)
	}()

	time.Sleep(3 * cfg.PollInterval)

	if err := os.Remove(path); err != nil {
		t.Fatalf("Remove(%q): %v", path, err)
	}
	time.Sleep(3 * cfg.PollInterval)

	mustWriteFile(t, path, "recreated\n")
	waitForContains(t, &out, path+": recreated")

	cancel()

	select {
	case err := <-errCh:
		require.True(t, err == nil || err == context.Canceled, "Run() returned error: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not exit after cancel")
	}
}

func TestScanPrunesDeletedFollowers(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	mustWriteFile(t, path, "seed\n")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := runner{
		ctx:       ctx,
		cfg:       Config{Dir: dir, Pattern: "*.log", PollInterval: 20 * time.Millisecond, FromStart: false},
		out:       io.Discard,
		events:    make(chan lineEvent),
		errs:      make(chan error, 1),
		followers: make(map[string]*follower),
	}

	if err := r.startFollower(path, false); err != nil {
		t.Fatalf("startFollower(%q): %v", path, err)
	}

	_, ok := r.followers[path]
	require.True(t, ok, "followers missing %q after start", path)

	require.NoError(t, os.Remove(path))

	require.NoError(t, r.scan(false))

	_, ok = r.followers[path]
	require.False(t, ok, "followers still contains deleted path %q", path)

	cancel()
	r.stopFollowers()
	r.wg.Wait()
}

func appendFile(t *testing.T, path, content string) {
	t.Helper()

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	require.NoError(t, err)
	defer f.Close()

	_, err = f.WriteString(content)
	require.NoError(t, err)
}

func waitForContains(t *testing.T, out *lockedBuffer, needle string) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if strings.Contains(out.String(), needle) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	require.Failf(t, "missing output", "output %q did not contain %q", out.String(), needle)
}

type lockedBuffer struct {
	mu sync.Mutex
	b  strings.Builder
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.b.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.b.String()
}
