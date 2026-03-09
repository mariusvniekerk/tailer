package tailer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/nxadm/tail"
)

type lineEvent struct {
	path string
	text string
}

type follower struct {
	tail *tail.Tail
}

type runner struct {
	ctx       context.Context
	cfg       Config
	out       io.Writer
	events    chan lineEvent
	errs      chan error
	followers map[string]*follower
	wg        sync.WaitGroup
}

// Run discovers matching files and writes filename-prefixed lines to out.
func Run(ctx context.Context, cfg Config, out io.Writer) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	r := runner{
		ctx:       ctx,
		cfg:       cfg,
		out:       out,
		events:    make(chan lineEvent),
		errs:      make(chan error, 1),
		followers: make(map[string]*follower),
	}

	return r.run()
}

func (r *runner) run() error {
	if err := r.scan(true); err != nil {
		r.stopFollowers()
		r.wg.Wait()
		return err
	}

	ticker := time.NewTicker(r.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			r.stopFollowers()
			r.wg.Wait()
			return r.ctx.Err()
		case err := <-r.errs:
			r.stopFollowers()
			r.wg.Wait()
			if errors.Is(err, context.Canceled) {
				return context.Canceled
			}
			return err
		case event := <-r.events:
			if _, err := fmt.Fprintf(r.out, "%s: %s\n", event.path, event.text); err != nil {
				r.stopFollowers()
				r.wg.Wait()
				return err
			}
		case <-ticker.C:
			if err := r.scan(false); err != nil {
				r.stopFollowers()
				r.wg.Wait()
				return err
			}
		}
	}
}

func (r *runner) startFollower(path string, fromStart bool) error {
	location, err := tailLocation(path, fromStart)
	if err != nil {
		return err
	}

	t, err := tail.TailFile(path, tail.Config{
		Location:      location,
		ReOpen:        true,
		MustExist:     false,
		Poll:          true,
		Follow:        true,
		CompleteLines: true,
		Logger:        tail.DiscardingLogger,
	})
	if err != nil {
		return err
	}

	r.followers[path] = &follower{tail: t}
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		defer t.Cleanup()

		for {
			select {
			case <-r.ctx.Done():
				return
			case line, ok := <-t.Lines:
				if !ok {
					return
				}
				if line.Err != nil {
					select {
					case r.errs <- fmt.Errorf("tail %s: %w", path, line.Err):
					default:
					}
					return
				}

				select {
				case r.events <- lineEvent{path: path, text: line.Text}:
				case <-r.ctx.Done():
					return
				}
			}
		}
	}()

	return nil
}

func (r *runner) scan(initial bool) error {
	paths, err := DiscoverFiles(r.cfg.Dir, r.cfg.Pattern)
	if err != nil {
		return err
	}
	r.pruneFollowers(paths)
	for _, path := range paths {
		if _, exists := r.followers[path]; exists {
			continue
		}
		fromStart := r.cfg.FromStart
		if !initial {
			fromStart = true
		}
		if err := r.startFollower(path, fromStart); err != nil {
			return err
		}
	}
	return nil
}

func (r *runner) pruneFollowers(paths []string) {
	visible := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		visible[path] = struct{}{}
	}

	for path, follower := range r.followers {
		if _, ok := visible[path]; ok {
			continue
		}
		_ = follower.tail.Stop()
		delete(r.followers, path)
	}
}

func (r *runner) stopFollowers() {
	for _, follower := range r.followers {
		_ = follower.tail.Stop()
	}
}
