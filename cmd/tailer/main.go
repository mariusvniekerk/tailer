package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mariusvniekerk/tailer/internal/tailer"
	"github.com/spf13/cobra"
)

func main() {
	cmd := newRootCmd(tailer.Run, os.Stdout, os.Stderr)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd(run func(context.Context, tailer.Config, io.Writer) error, stdout, stderr io.Writer) *cobra.Command {
	var cfg tailer.Config

	cmd := &cobra.Command{
		Use:   "tailer",
		Short: "Tail matching log files in a directory and auto-follow new ones",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.Validate(); err != nil {
				return err
			}

			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			if err := run(ctx, cfg, stdout); err != nil && err != context.Canceled {
				return err
			}

			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	flags := cmd.Flags()
	flags.StringVar(&cfg.Dir, "dir", ".", "Directory to watch for matching files.")
	flags.StringVar(&cfg.Pattern, "pattern", "*.log", "Glob pattern for files inside the watched directory.")
	flags.DurationVar(&cfg.PollInterval, "poll-interval", time.Second, "How often to rescan the directory for new files.")
	flags.BoolVar(&cfg.FromStart, "from-start", false, "Start existing files from byte 0 instead of the end.")

	return cmd
}
