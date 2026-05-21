package runtime

import _ "embed"

const OutputPath = "_sklair/runtime.mjs"

//go:embed runtime.mjs
var Module string
