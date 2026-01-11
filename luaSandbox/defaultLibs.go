package luaSandbox

import lua "github.com/yuin/gopher-lua"

type defaultLuaLib struct {
	libName string
	libFunc lua.LGFunction
}

var luaLibs = []defaultLuaLib{
	//{"package", lua.OpenPackage},
	{"", lua.OpenBase},
	{"table", lua.OpenTable},
	{"os", lua.OpenOs},
	{"string", lua.OpenString},
	{"math", lua.OpenMath},
	//{"coroutine", lua.OpenCoroutine}, // not much need for coroutines in hooks which are usually very sequential, but its here just in case (future?)
}

/*
func OpenPackage(L *LState) int {
	packagemod := L.RegisterModule(LoadLibName, loFuncs)

	L.SetField(packagemod, "preload", L.NewTable())

	loaders := L.CreateTable(len(loLoaders), 0)
	for i, loader := range loLoaders {
		L.RawSetInt(loaders, i+1, L.NewFunction(loader))
	}
	L.SetField(packagemod, "loaders", loaders)
	L.SetField(L.Get(RegistryIndex), "_LOADERS", loaders)

	loaded := L.NewTable()
	L.SetField(packagemod, "loaded", loaded)
	L.SetField(L.Get(RegistryIndex), "_LOADED", loaded)

	L.SetField(packagemod, "path", LString(loGetPath(LuaPath, LuaPathDefault)))
	L.SetField(packagemod, "cpath", emptyLString)

	L.SetField(packagemod, "config", LString(LuaDirSep+"\n"+LuaPathSep+
		"\n"+LuaPathMark+"\n"+LuaExecDir+"\n"+LuaIgMark+"\n"))

	L.Push(packagemod)
	return 1
}
*/

var remove = map[string][]string{
	"": {"_GOPHER_LUA_VERSION", "load", "dofile", "module", "loadfile", "loadstring", "setfenv", "require", "newproxy"},
	// TODO: modify package
	"os": {"execute", "exit", "getenv", "remove", "rename", "setenv" /*"setlocale",*/, "tmpname"},
}

func OpenSandboxedDefault(ls *lua.LState, opts *SandboxOptions) {
	for _, lib := range luaLibs {
		ls.Push(ls.NewFunction(lib.libFunc))
		ls.Push(lua.LString(lib.libName))
		ls.Call(1, 0)
	}

	for libName, funcs := range remove {
		var table lua.LValue

		if libName == "" {
			table = ls.Get(lua.GlobalsIndex)
		} else {
			table = ls.GetGlobal(libName)
		}

		tbl := table.(*lua.LTable)
		for _, funcName := range funcs {
			if libName == "os" && funcName == "exit" {
				tbl.RawSetString(funcName, ls.NewFunction(func(L *lua.LState) int {
					code := L.OptInt(1, 0)
					opts.ExitChannel <- code
					return 0
				}))

				continue
			}
			tbl.RawSetString(funcName, lua.LNil)
		}
	}

	//ls.SetGlobal("package", lua.LNil)
	// TODO: modify require() so that it is luvit-like, especially relative paths
	ls.SetGlobal("require", ls.NewFunction(func(L *lua.LState) int {
		L.RaiseError("require() is temporarily disabled in Sklair hooks for sandboxing reasons")
		return 0
	}))
}
