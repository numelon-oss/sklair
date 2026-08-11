{
  "description": "Create a HTML document and component, build them with Sklair, and inspect the resulting browser-native output."
}
---

# Your first build

Create `src/index.html` as a real document:

```html
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>My Sklair site</title>
</head>
<body>
  <SiteGreeting name="world"></SiteGreeting>
</body>
</html>
```

Then create `components/SiteGreeting.html`:

```html
<!doctype html>
<html>
<head></head>
<body>
  <h1>Hello, <span sklair-text="name"></span>.</h1>
</body>
</html>
```

The `sklair-text` binding implicitly declares the component's `name` prop. The invocation supplies it as an ordinary HTML attribute.

## Configure the directories

At the project root, write:

```json
{
  "input": "./src",
  "components": "./components",
  "layouts": "./layouts",
  "output": "./build"
}
```

An absent layouts directory is allowed. An absent components directory is not: component discovery needs the configured directory even when the first project barely deserves one.

## Compile the site

Run:

```sh
sklair build
```

The relevant part of `build/index.html` becomes:

```html
<body>
  <h1>Hello, <span>world</span>.</h1>
</body>
```

`SiteGreeting`, its prop, and the binding have disappeared. They were authoring structure, not an API the visitor's browser must download.

Sklair also adds its generator metadata and finalises the document head. Formatting may differ because the output is rendered from the parsed HTML tree; semantics are what the compiler preserves.

Change the prop, build again, mess around and inspect the output. Sklair can be understood even more when you look at what it *removes*.
