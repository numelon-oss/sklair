package snippets

import (
	_ "embed"
	"errors"
	"sklair/htmlUtilities"
	"strings"

	"golang.org/x/net/html"
)

//go:embed preventFOUC.html
var pfoucSrc string

// TODO: when building concurrency is added to sklair, then make GetFOUCNodes cache the result
// then use sync.Once
// var (
//
//	cachedStyle  *html.Node
//	cachedScript *html.Node
//	once         sync.Once
//
// )

func GetFOUCNodes(bgColour string) (*html.Node, *html.Node, error) {
	doc, err := html.Parse(strings.NewReader(pfoucSrc))
	if err != nil {
		return nil, nil, err
	}

	styleNode := htmlUtilities.FindTag(doc, "style")
	// styleNode = htmlUtilities.Clone(styleNode) // this is to be done by the caller
	if styleNode != nil && styleNode.FirstChild != nil && styleNode.FirstChild.Type == html.TextNode {
		styleNode.FirstChild.Data = strings.ReplaceAll(
			styleNode.FirstChild.Data,
			"#DEADBEEF", // placeholder
			bgColour,
		)
	}

	scriptNode := htmlUtilities.FindTag(doc, "script")

	if styleNode == nil || scriptNode == nil {
		return nil, nil, errors.New("could not find style or script nodes")
	}

	return styleNode, scriptNode, nil
}
