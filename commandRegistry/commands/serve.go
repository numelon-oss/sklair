package commands

import (
	"context"
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
	commandRegistry.Registry.Register(&commandRegistry.Command{
		Name:        "serve",
		Description: "Continuously builds and serves a Sklair project for development purposes",
		Configure: func(flags *flag.FlagSet) commandRegistry.Handler {
			var portN int
			var host string

			var open bool
			var autoRefresh bool
			var watch bool

			flags.IntVar(&portN, "port", 0, "Port to listen on")
			flags.StringVar(&host, "host", "localhost", "Host to listen on")

			flags.BoolVar(&open, "open", false, "Open the default browser") // TODO: actually use this flag value
			flags.BoolVar(&autoRefresh, "auto-refresh", true, "Automatically refresh connected preview instances")
			flags.BoolVar(&watch, "watch", true, "Watch for filesystem changes") // TODO: actually use this flag value

			return func(ctx context.Context, _ *commandRegistry.Environment, args []string) error {
				if len(args) != 0 {
					return commandRegistry.UsageErrorf("Serve does not accept positional arguments")
				}

				config, configDir, err := sklairConfig.LoadProjectConfig()
				if err != nil {
					return fmt.Errorf("load sklair.json: %w", err)
				}
				var rewriteValues []sklairConfig.ServeRewrite
				if config.Serve != nil {
					rewriteValues = config.Serve.Rewrites
				}
				rewrites, err := devserver.CompileRewrites(rewriteValues)
				if err != nil {
					return err
				}

				tmp, err := os.MkdirTemp("", "sklair-")
				if err != nil {
					return fmt.Errorf("create temporary directory: %w", err)
				}
				defer os.RemoveAll(tmp)

				listener, port, err := devserver.AcquirePort(host, portN)
				if err != nil {
					return fmt.Errorf("acquire port: %w", err)
				}
				defer listener.Close()

				wsThing := devserver.NewWS()

				// TODO: we need to be able to check whether the server started successfully in the first place or not
				// otherwise we are just walking in blind here
				// and dont know whether the file server is running or not
				// whilst still tracking the filesystem and recompiling every time...
				go devserver.Serve(listener, tmp, port, wsThing, rewrites)

				err = building.Build(config, configDir, tmp)
				if err != nil {
					return fmt.Errorf("build project: %w", err)
				}

				// for now: ENTIRE project is rebuild on change
				// but in the future maybe only rebuild changed files: see comment at very top

				watchPaths, err := getWatchPaths(config.Input, config.Components, config.Layouts)
				if err != nil {
					return fmt.Errorf("resolve watch paths: %w", err)
				}

				events, errs := devserver.Watch(watchPaths...)

				for {
					select {
					case <-ctx.Done():
						return nil
					case <-events:
						//_ = os.RemoveAll(tmp)
						//_ = os.MkdirAll(tmp, 0755)

						err = building.Build(config, configDir, tmp)
						if err != nil {
							return fmt.Errorf("rebuild project: %w", err)
						}

						if autoRefresh {
							wsThing.Send <- "reload"
						}
					case err := <-errs:
						logger.Error("%v", err)
					}
				}
			}
		},
	})
}
