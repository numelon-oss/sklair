local arr = require("../util/efficientArray")

-- void elements are self closing tags (implied close even if close tag not present)
-- https://developer.mozilla.org/en-US/docs/Glossary/Void_element
return arr {
    "area",
    "base",
    "br",
    "col",
    "embed",
    "hr",
    "img",
    "input",
    "link",
    "meta",
    "param",
    "source",
    "track",
    "wbr"
}
