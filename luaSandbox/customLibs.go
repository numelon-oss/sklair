package luaSandbox

import (
	lua "github.com/yuin/gopher-lua"
	json "layeh.com/gopher-json"
)

type LFuncWithSandboxContext func(opts *SandboxOptions) lua.LGFunction

type customLuaLib struct {
	libName    string
	libFactory LFuncWithSandboxContext
}

var customLibs = []customLuaLib{
	{"fs", openFs},
	{"json", func(_ *SandboxOptions) lua.LGFunction {
		return json.Loader
	}},
}

func OpenSandboxedCustom(ls *lua.LState, opts *SandboxOptions) {
	for _, lib := range customLibs {
		loader := lib.libFactory(opts)

		ls.Push(ls.NewFunction(loader))
		ls.Push(lua.LString(lib.libName))
		ls.Call(1, 0)
	}
}
