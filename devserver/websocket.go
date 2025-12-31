package devserver

import (
	"sklair/logger"
	"sync"

	"golang.org/x/net/websocket"
)

const WSPath = "_sklair/ws"

type WS struct {
	clients map[*websocket.Conn]struct{}
	mu      sync.Mutex
	Send    chan string
}

func NewWS() *WS {
	ws := &WS{
		clients: make(map[*websocket.Conn]struct{}),
		Send:    make(chan string),
	}
	go ws.run()
	return ws
}

func (ws *WS) run() {
	for msg := range ws.Send {
		ws.mu.Lock()
		for client := range ws.clients {
			_ = websocket.Message.Send(client, msg)
			logger.Debug("Sent message to client")
		}
		ws.mu.Unlock()
	}
}

func (ws *WS) HandleWS(c *websocket.Conn) {
	ws.mu.Lock()
	ws.clients[c] = struct{}{}
	ws.mu.Unlock()

	defer func() {
		ws.mu.Lock()
		delete(ws.clients, c)
		ws.mu.Unlock()
		_ = c.Close()
		//logger.Debug("Client disconnected")
	}()

	logger.Debug("New client connected")

	// just block until the client disconnects
	var msg string
	for {
		if err := websocket.Message.Receive(c, &msg); err != nil {
			return
		}
	}
}
