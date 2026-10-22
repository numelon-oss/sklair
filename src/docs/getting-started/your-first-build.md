{
  "description": "Create one HTML component, use it from a page, and inspect the HTML produced by Sklair."
}
---

# Your first build

The project needs a configuration file, one page, and one [component](../glossary/component.md).

## Configure Sklair

Write this to `sklair.json`:

```json title="sklair.json" linenums="1"
{
  "$schema": "https://github.com/numelon-oss/sklair/releases/latest/download/sklair.schema.json",
  "input": "./src",
  "components": "./components",
  "layouts": "./layouts",
  "output": "./build"
}
```

The `$schema` line is optional. Editors which support JSON Schema can use it for completion and validation.

## Create the component

Write this to `components/SiteGreeting.html`:

```html title="components/SiteGreeting.html" linenums="1" hl_lines="7"
<!doctype html>
<html>
<head></head>
<body>
  <h1>
    Hello,
    <span sklair-text="name"></span>.
  </h1>
</body>
</html>
```

`#!html sklair-text="name"` is a [binding](../glossary/binding.md). It inserts the `name` [prop](../glossary/prop.md) as text and removes the `sklair-text` attribute from the result.

## Use the component

Write this to `src/index.html`:

```html title="src/index.html" linenums="1" hl_lines="8"
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

The `#!html <SiteGreeting>` tag is a [component use](../glossary/component-use.md). Its `name` attribute supplies the value read by the binding.

## Compile it

Run Sklair from the directory containing `sklair.json`:

```bash title="Terminal"
sklair build
```

Open `build/index.html`. Its body contains ordinary HTML:

```html title="build/index.html"
<body>
  <h1>
    Hello,
    <span>world</span>.
  </h1>
</body>
```

The component tag, prop, and binding are gone. Change `name`, build again, mess around and inspect the output. Sklair is easiest to understand when you see what it removes.
