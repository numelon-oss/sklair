package devserver

import (
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sklair/logger"
	"strconv"

	"golang.org/x/net/websocket"
)

func AcquirePort(addr string, port int) (net.Listener, int, error) {
	var listener net.Listener
	var err error

	if port == 0 {
		// if the provided port is 0, then we simply try to find a free port between 8080 and 8090
		for i := 0; i < 11; i++ {
			port = 8080 + i

			logger.Debug("trying port %d", port)
			listener, err = net.Listen("tcp", addr+":"+strconv.Itoa(port))
			if err == nil {
				break
			}
		}

		if listener == nil {
			return nil, 0, errors.New("no free ports in the range 8080-8090")
		}
	} else {
		// otherwise, an explicit port was provided so TRY to bind to it regularly
		listener, err = net.Listen("tcp", addr+":"+strconv.Itoa(port))
		if err != nil {
			return nil, 0, err
		}
	}

	return listener, port, nil
}

func try404(root string, w http.ResponseWriter, r *http.Request) {
	p := filepath.Join(root, "404.html")
	if _, err := os.Stat(p); err == nil {
		w.WriteHeader(http.StatusNotFound)
		http.ServeFile(w, r, p)
		return
	}

	p = filepath.Join(root, "404.shtml")
	if _, err := os.Stat(p); err == nil {
		w.WriteHeader(http.StatusNotFound)
		http.ServeFile(w, r, p)
		return
	}

	http.Error(w, "404 not found", http.StatusNotFound)
}

func Serve(listener net.Listener, tmp string, port int, wsThing *WS) {
	staticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("cache-control", "no-cache, no-store, must-revalidate")

		path := filepath.Join(tmp, filepath.Clean(r.URL.Path))
		if file, err := os.Stat(path); err == nil && file.IsDir() {
			path = filepath.Join(path, "index.html")
		}

		if _, err := os.Stat(path); err == nil {
			http.ServeFile(w, r, path)
			return
		}

		try404(tmp, w, r)
	})

	mux := http.NewServeMux()
	mux.Handle("/"+WSPath, websocket.Handler(wsThing.HandleWS))
	mux.Handle("/", staticHandler)

	logger.Info("Will be listening on http://localhost:" + strconv.Itoa(port) + "/")
	if err := http.Serve(listener, mux); err != nil {
		logger.Error(err.Error())
	}
}
