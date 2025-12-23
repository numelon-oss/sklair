package snippets

import "golang.org/x/net/html"

func newMetaNode(name, property, content string) *html.Node {
	key := "name"
	val := name
	if property != "" {
		key = "property"
		val = property
	}

	return &html.Node{
		Type: html.ElementNode,
		Data: "meta",
		Attr: []html.Attribute{
			{Key: key, Val: val},
			{Key: "content", Val: content},
		},
	}
}

// TODO: extend this MASSIVELY so that it includes every single property possible from BOTH opengraph and twitter
// https://ogp.me/
// https://developer.x.com/en/docs/x-for-websites/cards/guides/getting-started
/*
<OpenGraph
  title="Sklair | HTML deserved better."
  description="Sklair is a modern compiler…"
  image="https://sklair.numelon.com/img/opengraph.jpg"
  url="https://sklair.numelon.com"
  siteName="Sklair"
  twitterSite="@username" <-- THIS isnt implemented in the function below!
  twitterCreator="@username" <-- THIS isnt implemented in the function below!
								basically, just add a bunch of options from both standards (optional attributes to this opengraph component) so that its super duper customisable
  type="website"
  imageSize="large"
/>

*/

// TODO: all opengraph tags should be inserted at the very end of the head tag of outputs
// after inserting everything, at the end of the for loop of toReplace, go through the head and do a heuristic sort
func OpenGraph(originalTag *html.Node) []*html.Node {
	var out []*html.Node

	var (
		siteName    string
		title       string
		description string
		image       string
		url         string
		ogType      = "website" // default
		imageSize   = "large"   // default
	)

	for _, attr := range originalTag.Attr {
		switch attr.Key {
		case "site_name":
			siteName = attr.Val
		case "title":
			title = attr.Val
		case "description":
			description = attr.Val
		case "image":
			image = attr.Val
		case "url":
			url = attr.Val
		case "type":
			ogType = attr.Val
		case "image_size":
			imageSize = attr.Val
			//if attr.Val == "small" {
			//	imageSize = "small"
			//}
		}
	}

	if siteName != "" {
		out = append(out, newMetaNode("", "og:site_name", siteName))
	}

	if title != "" {
		out = append(out,
			newMetaNode("twitter:title", "", title),
			newMetaNode("", "og:title", title),
		)
	}

	if description != "" {
		out = append(out,
			newMetaNode("description", "", description),
			newMetaNode("twitter:description", "", description),
			newMetaNode("", "og:description", description),
		)
	}

	// TODO: always emit twitter card if image exists
	if image != "" {
		out = append(out,
			newMetaNode("twitter:image", "", image),
			newMetaNode("", "og:image", image),
		)

		card := "summary_large_image"
		if imageSize == "small" {
			card = "summary"
		}
		out = append(out, newMetaNode("twitter:card", "", card))
	}

	if url != "" {
		out = append(out,
			newMetaNode("twitter:url", "", url),
			newMetaNode("", "og:url", url),
		)
	}

	if ogType != "" {
		out = append(out,
			newMetaNode("", "og:type", ogType),
		)
	}

	return out
}
