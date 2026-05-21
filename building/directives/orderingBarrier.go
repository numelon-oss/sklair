package directives

import (
	"strings"

	"golang.org/x/net/html"
)

// TODO: note in the documentation that items inside of ordering barriers will not be deduplicated
// or maybe if ordering barriers are present with the same content then they will be deduplicated
// basically just treat head segment as one item and hash it, deduplicate head segments

// OB = ordering barrier

func IsOBStart(n *html.Node) (bool, string) {
	if n.Type != html.CommentNode {
		return false, ""
	}

	text := strings.TrimSpace(n.Data)

	// don't treat the end as the start accidentally, I already made that mistake
	if text == "sklair:ordering-barrier-end" {
		return false, ""
	}

	if !strings.HasPrefix(text, "sklair:ordering-barrier ") {
		return false, ""
	}

	var treatAsTag string

	parts := strings.FieldsSeq(text)
	for part := range parts {
		if after, ok := strings.CutPrefix(part, "treat-as="); ok {
			treatAsTag = after
		}
	}

	return true, treatAsTag
}

func IsOBEnd(n *html.Node) bool {
	return n.Type == html.CommentNode && strings.TrimSpace(n.Data) == "sklair:ordering-barrier-end"
}
