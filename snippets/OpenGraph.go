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

	for _, attr := range originalTag.Attr {
		if attr.Key == "site_name" {
			out = append(out, newMetaNode("", "og:site_name", attr.Val))
		} else if attr.Key == "title" {
			//out = append(out, newMetaNode("title", "", attr.Val)) // TODO: useless
			out = append(out, newMetaNode("twitter:title", "", attr.Val))
			out = append(out, newMetaNode("", "og:title", attr.Val))
		} else if attr.Key == "description" {
			out = append(out, newMetaNode("description", "", attr.Val))
			out = append(out, newMetaNode("twitter:description", "", attr.Val))
			out = append(out, newMetaNode("", "og:description", attr.Val))
		} else if attr.Key == "image" {
			out = append(out, newMetaNode("twitter:image", "", attr.Val))
			out = append(out, newMetaNode("", "og:image", attr.Val))
		} else if attr.Key == "url" {
			out = append(out, newMetaNode("twitter:url", "", attr.Val))
			out = append(out, newMetaNode("", "og:url", attr.Val))
		} else if attr.Key == "type" {
			out = append(out, newMetaNode("", "og:type", attr.Val))
		} else if attr.Key == "image_size" {
			val := "summary_large_image"
			if attr.Val == "small" {
				val = "summary"
			}
			out = append(out, newMetaNode("twitter:card", "", val))
		}
	}

	return out
}
