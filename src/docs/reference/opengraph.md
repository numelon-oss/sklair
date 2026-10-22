{
  "description": "Use Sklair's built-in OpenGraph component to generate page descriptions and social-sharing metadata."
}
---

# OpenGraph component

OpenGraph metadata gives social platforms and messaging apps a title, description, and image to use when someone shares a link to your page. Sklair's built-in component generates OpenGraph and Twitter card tags, so you can define rich link previews in one place.

Sklair includes a native `OpenGraph` component for this. No component file or configuration is needed. Place it in the document body, or inside an ordinary component's body; Sklair removes the tag and contributes its metadata to the final head.

```html title="Page metadata"
<OpenGraph
  site_name="Example"
  title="Example page"
  description="A short description of this page."
  image="https://example.com/social.jpg"
  url="https://example.com/page/"
  type="website"
  image_size="large"
></OpenGraph>
```

## Attributes

| Attribute | Default | Output |
| --- | --- | --- |
| `site_name` | empty | `og:site_name` |
| `title` | empty | `og:title` and `twitter:title` |
| `description` | empty | Ordinary `description` metadata, `og:description`, and `twitter:description` |
| `image` | empty | `og:image`, `twitter:image`, and `twitter:card` |
| `url` | empty | `og:url` and `twitter:url` |
| `type` | `website` | `og:type` |
| `image_size` | `large` | Selects the Twitter card type when an image is supplied |

Use the underscore spellings `site_name` and `image_size`. Other attribute names are ignored. Omitted or empty values produce no corresponding metadata, except that an omitted `type` defaults to `website`.

`image_size="small"` selects `summary`. All other values select `summary_large_image`. No Twitter card tag is emitted without an image.

Use full public URLs for `image` and `url`. The component copies their values into metadata; it does not resolve paths, upload an image, or generate a social image.

## Generated head

The component generates only metadata. It does not set the document's `<title>` element or its canonical link; author those separately when needed.

The metadata is grouped in an ordering barrier with `treat-as=dont-care`. Nodes in that group are not individually deduplicated, so use one `OpenGraph` tag per document and avoid also emitting the same metadata elsewhere. See [Head processing](../sklair/head-processing.md).

You can use bindings or inline Lua in an ordinary component to supply its values. As with other Sklair components, use an explicit closing tag.
