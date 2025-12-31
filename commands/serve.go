package commands

import (
	"os"
	"sklair/building"
	"sklair/commandRegistry"
	"sklair/devserver"
	"sklair/logger"
	"sklair/sklairConfig"
	"strconv"
	"strings"
)

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
		Aliases:     []string{"s"},
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

			listener, port, err := devserver.AcquirePort("localhost", 0)
			if err != nil {
				logger.Error("could not acquire port : %s", err.Error())
				return 1
			}
			defer listener.Close()

			addr := "localhost:" + strconv.Itoa(port)
			devserver.WSDevScript = strings.ReplaceAll(devserver.WSDevScript, "WEBSOCKET_PATH", devserver.WSPath)
			devserver.WSDevScript = strings.ReplaceAll(devserver.WSDevScript, "WEBSOCKET", addr)

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

			// TODO: add port flag, auto_refresh bool (websocket) flag
			// track changes from the following directories:
			// - source directory (excluding components dir, if it is within the source directory)
			// OR if the components directory is within the source directory then just ONLY track the source directory anyways
			// - components directory by itself
			// from all tracked directories, output dir must be excluded along with common excluded directories
			// for now: ENTIRE project is rebuild on change

			// but in the future maybe only rebuild changed files: see comment at very top

			// try all ports from 8080 upwards (but obviously at some point theres a limit)
			// websocket lives on same http, just connection upgrade
			// after decided, they are now just hardcoded

			events, errs := devserver.Watch(config.Input)

			for {
				select {
				case <-events:
					_ = os.RemoveAll(tmp)
					_ = os.MkdirAll(tmp, 0755)

					err = building.Build(config, configDir, tmp)
					if err != nil {
						logger.Error(err.Error())
						return 1
					}

					wsThing.Send <- "reload"
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
