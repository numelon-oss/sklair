package cli

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"sklair/building"
	"sklair/devserver"
	"sklair/sklairConfig"
	"sklair/util"

	commandregistry "github.com/numelon-oss/go-command-registry"
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

		// paths beneath input are already covered by its recursive watcher, hence its not needed to separately include components etc if inside src
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
func serveCommand() *Command {
	return &Command{
		Name:        "serve",
		Description: "Build and serve a project for development",
		Arguments:   "[options]",
		Examples:    []string{"sklair serve", "sklair serve --port 8080 --open", "sklair serve --watch=false"},
		Configure: func(flags *flag.FlagSet) Handler {
			host := flags.String("host", "localhost", "Listening address (default localhost)")
			port := flags.Int("port", 0, "Port; zero selects an available port between 8080 and 8090")
			open := flags.Bool("open", false, "Open the preview in your default browser after the first successful build")
			autoRefresh := flags.Bool("auto-refresh", true, "Refresh connected pages after rebuilding (default true)")
			watch := flags.Bool("watch", true, "Watch source files and rebuild on change (default true)")
			return func(ctx context.Context, _ struct{}, args []string) error {
				if len(args) != 0 {
					return commandregistry.UsageErrorf("serve does not accept positional arguments")
				}
				if *port < 0 || *port > 65535 {
					return commandregistry.UsageErrorf("port must be between 0 and 65535")
				}
				return runServe(ctx, *host, *port, *open, *autoRefresh, *watch)
			}
		},
	}
}

func runServe(ctx context.Context, host string, port int, open, autoRefresh, watch bool) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()
	config, configDir, err := sklairConfig.LoadProjectConfig()
	if err != nil {
		return fmt.Errorf("could not load sklair.json: %w", err)
	}
	tmp, err := os.MkdirTemp("", "sklair-")
	if err != nil {
		return fmt.Errorf("could not create temporary directory: %w", err)
	}
	defer os.RemoveAll(tmp)
	listener, port, err := devserver.AcquirePort(host, port)
	if err != nil {
		return fmt.Errorf("could not acquire port: %w", err)
	}
	defer listener.Close()
	if err := building.Build(config, configDir, tmp); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return nil
	}

	var events <-chan bool
	var watchErrors <-chan error
	if watch {
		watchPaths, err := getWatchPaths(
			filepath.Join(configDir, config.Input),
			filepath.Join(configDir, config.Components),
			filepath.Join(configDir, config.Layouts),
		)
		if err != nil {
			return err
		}
		events, watchErrors = devserver.Watch(ctx, watchPaths...)
	}
	ws := devserver.NewWS()
	defer close(ws.Send)
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- devserver.Serve(listener, tmp, ws) }()
	listenAddress := net.JoinHostPort(strings.Trim(host, "[]"), strconv.Itoa(port))
	if host == "" {
		listenAddress = listener.Addr().String()
	}
	logger.Info("Listening on http://%s/", listenAddress)
	if open {
		// fyi: this (|0.0.0.0|::) fallback to localhost only affects what is auto opened in the local browser
		// not what sklair serve actually listens on
		browserHost := strings.Trim(host, "[]")
		if browserHost == "" || browserHost == "0.0.0.0" || browserHost == "::" {
			browserHost = "localhost"
		}
		browserURL := "http://" + net.JoinHostPort(browserHost, strconv.Itoa(port)) + "/"
		logger.Info("Opening local preview at %s", browserURL)
		if err := util.OpenBrowser(browserURL); err != nil {
			logger.Warning("Could not open browser: %s. Preview available at %s", err.Error(), browserURL)
		}
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-serverErrors:
			return fmt.Errorf("development server stopped: %w", err)
		case _, ok := <-events:
			if !ok {
				events = nil
				continue
			}
			// TODO: rebuild only changed documents and assets once dependency tracking is available.
			if err := building.Build(config, configDir, tmp); err != nil {
				return err
			}
			if autoRefresh {
				select {
				case ws.Send <- "reload":
				case <-ctx.Done():
					return nil
				}
			}
		case err, ok := <-watchErrors:
			if !ok {
				watchErrors = nil
				continue
			}
			return fmt.Errorf("could not watch project files: %w", err)
		}
	}
}
