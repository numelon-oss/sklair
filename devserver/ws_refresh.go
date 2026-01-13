package devserver

import (
	_ "embed"
	"strings"

	"golang.org/x/net/html"
)

// WSDevScript is the source of the script that is saved to _sklair/sklair_ws.js in the build directory when the dev server is enabled
//
//go:embed ws_refresh.js
var WSDevScript string

const WSDevScriptPath = "/_sklair/ws_refresh.js"

var WSScriptNode = &html.Node{
	Type: html.ElementNode,
	Data: "script",
	Attr: []html.Attribute{{Key: "src", Val: WSDevScriptPath}},
}

func init() {
	WSDevScript = strings.ReplaceAll(WSDevScript, "WEBSOCKET_PATH", WSPath)
}
