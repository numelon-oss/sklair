local arr = require("../util/efficientArray")

-- https://devdoc.net/web/developer.mozilla.org/en-US/docs/HTML/Inline_elements.html
return arr {
    "a",
    "b",
    "big",
    "i",
    "small",
    "tt",
    "abbr",
    "acronym",
    "cite",
    "code",
    "dfn",
    "em",
    "kbd",
    "strong",
    "samp",
    "time",
    "var",
    "bdo",
    "br",
    "img",
    "map",
    "object",
    "q",
    "script",
    "span",
    "sub",
    "sup",
    "button",
    "input",
    "label",
    "select",
    "textarea",

    -- TODO: ???
    "u",
    "mark",

    -- TODO: separate some into separate raw text tags instead of just inline
}
