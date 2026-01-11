package devserver

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

func debounce(in <-chan bool, delay time.Duration) <-chan bool {
	out := make(chan bool)

	go func() {
		defer close(out)

		var (
			timer   *time.Timer
			pending bool
		)

		for {
			var timerC <-chan time.Time
			if timer != nil {
				timerC = timer.C
			}

			select {
			case _, ok := <-in:
				if !ok {
					return
				}

				pending = true

				if timer == nil {
					timer = time.NewTimer(delay)
				} else {
					if !timer.Stop() {
						<-timer.C
					}
					timer.Reset(delay)
				}

			case <-timerC:
				if pending {
					out <- true
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
// - source directory (excluding components dir, if it is within the source directory)
// OR if the components directory is within the source directory then just ONLY track the source directory anyways
// - components directory by itself
// from all tracked directories, output dir must be excluded along with common excluded directories

// TODO: dir parameter removed in favour of source and components dir and excludes list (when above changes are implemented)
// also refer to commands/serve.go for more information
func Watch(dir string) (<-chan bool, <-chan error) {
	events := make(chan bool)
	errs := make(chan error)

	go func() {
		defer close(events)
		defer close(errs)

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			errs <- err
			return
		}
		defer watcher.Close()

		// recursively watch ALL subdirectories
		err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
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
			errs <- err
		}

		for {
			select {
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
					events <- true
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}

				errs <- err
			}
		}
	}()

	return debounce(events, 150*time.Millisecond), errs
}
