package luaSandbox

import (
	lua "github.com/yuin/gopher-lua"
	json "layeh.com/gopher-json"
)

type customLuaLib struct {
	libName string
	loader  lua.LGFunction
}

type HookMode uint8

const (
	HookModePre HookMode = iota
	HookModePost
)

var customLibs = []customLuaLib{
	//{"http", openHttp},
	{"json", func(L *lua.LState) int {
		n := json.Loader(L)
		mod := L.Get(-1)
		L.SetGlobal("json", mod)
		L.Pop(n)
		return 0
	}},
}

func OpenSandboxedCustom(ls *lua.LState) {
	for _, lib := range customLibs {
		ls.Push(ls.NewFunction(lib.loader))
		ls.Push(lua.LString(lib.libName))
		ls.Call(1, 0)
	}
}
