package devserver

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

func debounce(ctx context.Context, in <-chan bool, delay time.Duration) <-chan bool {
	out := make(chan bool)

	go func() {
		defer close(out)

		var (
			timer   *time.Timer
			pending bool
		)
		defer func() {
			if timer != nil {
				timer.Stop()
			}
		}()

		for {
			var timerC <-chan time.Time
			if timer != nil {
				timerC = timer.C
			}

			select {
			case <-ctx.Done():
				return
			case _, ok := <-in:
				if !ok {
					return
				}

				pending = true

				if timer == nil {
					timer = time.NewTimer(delay)
				} else {
					if !timer.Stop() {
						select {
						case <-timer.C:
						default:
						}
					}
					timer.Reset(delay)
				}

			case <-timerC:
				if pending {
					select {
					case out <- true:
					case <-ctx.Done():
						return
					}
					pending = false
				}
				timer = nil
			}
		}
	}()

	return out
}

func shouldWatch(path string) bool {
	base := filepath.Base(path)
	switch base {
	case ".sklair", ".git":
		return false
	}

	return true
}

// track changes from the following directories:
// - source directory
// - component and layout directories which are not already within source
// from all tracked directories, output dir must be excluded along with common excluded directories

// TODO: paths should eventually be derived from resolved build inputs and excludes
// also refer to cli/serve.go for more information
func Watch(ctx context.Context, paths ...string) (<-chan bool, <-chan error) {
	events := make(chan bool)
	errs := make(chan error)

	go func() {
		defer close(events)
		defer close(errs)

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			select {
			case errs <- err:
			case <-ctx.Done():
			}
			return
		}
		defer watcher.Close()

		for _, root := range paths {
			// recursively watch ALL subdirectories
			err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					if shouldWatch(path) {
						return watcher.Add(path)
					} else {
						return filepath.SkipDir
					}
				}

				return nil
			})
			if err != nil {
				select {
				case errs <- err:
				case <-ctx.Done():
				}
				return
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			case e, ok := <-watcher.Events:
				if !ok {
					return
				}

				// handle new directories dynamically
				if e.Op&fsnotify.Create != 0 {
					if info, err := os.Stat(e.Name); err == nil && info.IsDir() && shouldWatch(e.Name) {
						_ = watcher.Add(e.Name)
					}
				}

				// we only want writes, creates and deletes
				if e.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
					select {
					case events <- true:
					case <-ctx.Done():
						return
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}

				select {
				case errs <- err:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return debounce(ctx, events, 150*time.Millisecond), errs
}
