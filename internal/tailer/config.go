package tailer

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// Config controls one tailer process.
type Config struct {
	Dir          string
	Pattern      string
	PollInterval time.Duration
	FromStart    bool
}

// Validate rejects unusable runtime configuration early.
func (c Config) Validate() error {
	switch {
	case c.Dir == "":
		return errors.New("dir is required")
	case c.Pattern == "":
		return errors.New("pattern is required")
	case c.PollInterval <= 0:
		return errors.New("poll interval must be positive")
	}

	info, err := os.Stat(c.Dir)
	if err != nil {
		return fmt.Errorf("stat dir: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("dir %q is not a directory", c.Dir)
	}

	return nil
}
