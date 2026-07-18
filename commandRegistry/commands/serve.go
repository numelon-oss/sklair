package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sklair/building"
	"sklair/commandRegistry"
	"sklair/devserver"
	"sklair/sklairConfig"
	"strings"

	"github.com/numelon-oss/go-logger"
)

func getWatchPaths(src string, extraPaths ...string) ([]string, error) {
	watchPaths := []string{src}

	inputAbs, err := filepath.Abs(src)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path of %s: %w", src, err)
	}

	for _, path := range extraPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, err
		}

		absolute, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve absolute path of %s: %w", path, err)
		}
		relative, err := filepath.Rel(inputAbs, absolute)

		// paths already beneath input are covered by its recursive watcher
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			watchPaths = append(watchPaths, path)
		}
	}

	return watchPaths, nil
}

// REBUILDING ONLY CHANGES FILES:
// in order to do so, we track changes from source dir and component dir
// if the change is from source dir, then rebuild only the singular HTML file
// if the change is from components dir, then rebuild all HTML files which use that component
// this however requires dependency tracking, which will be implemented later only
// so for now the entire project gets rebuilt

// however...
// TODO: on each rebuild, do not re-copy static files. only copy new static files if they are changed
// this will save a lot of time (bc no need to copy static files every time)
// therefore ONLY process (build) changed HtmlFiles, not StaticFiles
// but this still requires a bit of work but its much easier than the former

// TODO: add the following flags
// port (default is 8080 upwards)
// host (default is 127.0.0.1)
// --open (opens browser)
// --auto_refresh=true|false (websocket control, default true)
// --watch=true|false (watch for changes, default true)
func init() {
	var portN int
	var host string

	var open bool
	var autoRefresh bool
	var watch bool

	commandRegistry.Registry.Register(&commandRegistry.Command{
		Name:        "serve",
		Description: "Continuously builds and serves a Sklair project for development purposes",

		Flags: func() *flag.FlagSet {
			fs := flag.NewFlagSet("serve", flag.ContinueOnError)

			fs.IntVar(&portN, "port", 0, "Port to listen on")
			fs.StringVar(&host, "host", "localhost", "Host to listen on")

			fs.BoolVar(&open, "open", false, "Open the default browser") // TODO: actually use this flag value
			fs.BoolVar(&autoRefresh, "auto-refresh", true, "Automatically refresh connected preview instances")
			fs.BoolVar(&watch, "watch", true, "Watch for filesystem changes") // TODO: actually use this flag value

			return fs
		},

		Run: func(args []string) int {
			config, configDir, err := sklairConfig.LoadProjectConfig()
			if err != nil {
				logger.Error("could not load sklair.json : %s", err.Error())
				return 1
			}

			tmp, err := os.MkdirTemp("", "sklair-")
			if err != nil {
				logger.Error("could not create temporary directory : %s", err.Error())
				return 1
			}
			defer os.RemoveAll(tmp)

			listener, port, err := devserver.AcquirePort(host, portN)
			if err != nil {
				logger.Error("could not acquire port : %s", err.Error())
				return 1
			}
			defer listener.Close()

			wsThing := devserver.NewWS()

			// TODO: we need to be able to check whether the server started successfully in the first place or not
			// otherwise we are just walking in blind here
			// and dont know whether the file server is running or not
			// whilst still tracking the filesystem and recompiling every time...
			go devserver.Serve(listener, tmp, port, wsThing)

			err = building.Build(config, configDir, tmp)
			if err != nil {
				logger.Error(err.Error())
				return 1
			}

			// for now: ENTIRE project is rebuild on change
			// but in the future maybe only rebuild changed files: see comment at very top

			watchPaths, err := getWatchPaths(config.Input, config.Components, config.Layouts)
			if err != nil {
				logger.Error(err.Error())
				return 1
			}

			events, errs := devserver.Watch(watchPaths...)

			for {
				select {
				case <-events:
					//_ = os.RemoveAll(tmp)
					//_ = os.MkdirAll(tmp, 0755)

					err = building.Build(config, configDir, tmp)
					if err != nil {
						logger.Error(err.Error())
						return 1
					}

					if autoRefresh {
						wsThing.Send <- "reload"
					}
				case err := <-errs:
					logger.Error(err.Error())
				}
			}

			// TODO: add a channel which is used for receiving Ctrl+C signals for graceful shutdown,
			// perhaps supply that channel to the Watch function to make all the defers run

			return 0
		},
	})
}
