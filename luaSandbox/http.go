package luaSandbox

type HttpContext struct {
	Mode HookMode

	HttpAllowed    bool // whether http is allowed, as opposed to https
	AllowedHosts   []string
	AllowedMethods []string

	MaxResponseBytes    int64
	TimeoutMilliseconds int
	FollowRedirects     bool
	MaxRedirects        int
}

//
//func openHttp(opts *SandboxOptions) lua.LGFunction {
//	return func(L *lua.LState) int {
//		contextualised := make(map[string]lua.LGFunction, len(fsFuncs))
//		for name, f := range httpFuncs {
//			contextualised[name] = f(&opts.FSContext)
//		}
//
//		fsMod := L.RegisterModule("http", contextualised)
//		L.Push(fsMod)
//		return 0
//	}
//}
//
//type lFuncWithHttpContext func(*FSContext) lua.LGFunction
//
//var httpFuncs = map[string]lFuncWithHttpContext{
//	"read":    readFile,
//	"write":   writeFile,
//	"scandir": scanDir,
//}
//
//func _resolvePath(ctx *FSContext, path string, mode AccessMode) (string, error) {
//	// TODO: what if theres a symlink in the parent and suddenly were exposed now?
//	// yikes, but good for now because hooks are assumed trusted for now
//	// TODO: later use filepath.Clean, filepath.IsAbs, filepath.Rel for full canonical validation against the above issue
//	if strings.Contains(path, "..") && !(strings.Count(path, "..") == 1 && strings.HasPrefix(path, "project:../")) {
//		return "", errors.New("path traversal is not allowed")
//	}
//
//	switch {
//	case strings.HasPrefix(path, "cache:"):
//		return filepath.Join(ctx.CacheDir, strings.TrimPrefix(path, "cache:")), nil
//
//	case strings.HasPrefix(path, "project:"):
//		if mode != AccessModeRead {
//			return "", errors.New("project files are read-only")
//		}
//		return filepath.Join(ctx.ProjectDir, strings.TrimPrefix(path, "project:")), nil
//
//	case strings.HasPrefix(path, "temp:"):
//		return filepath.Join(ctx.TempDir, strings.TrimPrefix(path, "temp:")), nil
//
//	case strings.HasPrefix(path, "generated:"):
//		return filepath.Join(ctx.GeneratedDir, strings.TrimPrefix(path, "generated:")), nil
//
//	case strings.HasPrefix(path, "built:"):
//		if ctx.Mode != HookModePost {
//			return "", errors.New("built files are only available in post-build hooks")
//		}
//		return filepath.Join(ctx.BuiltDir, strings.TrimPrefix(path, "built:")), nil
//	}
//
//	return "", errors.New("path must start with `cache`, `project`, `temporary`, `generated`, or `built`, followed by a colon and a relative path")
//}
//
//// readFile reads the contents of a file specified by the first argument and returns the data and a potential error.
//// on success, returns the data as a string. on error, returns nil and the error message.
//func _readFile(ctx *FSContext) lua.LGFunction {
//	return func(L *lua.LState) int {
//		name := L.CheckString(1)
//
//		path, err := resolvePath(ctx, name, AccessModeRead)
//		if err != nil {
//			L.Push(lua.LNil)
//			L.Push(lua.LString(err.Error()))
//			return 2
//		}
//
//		data, err := os.ReadFile(path)
//		if err != nil {
//			L.Push(lua.LNil)
//			L.Push(lua.LString(err.Error()))
//			return 2
//		}
//
//		L.Push(lua.LString(data))
//		return 1
//	}
//}
//
//// writeFile writes the contents of the second argument to the file specified by the first argument.
//// on success, returns true. on error, returns the nil and the error message.
//func _writeFile(ctx *FSContext) lua.LGFunction {
//	return func(L *lua.LState) int {
//		name := L.CheckString(1)
//		data := L.CheckString(2)
//
//		path, err := resolvePath(ctx, name, AccessModeWrite)
//		if err != nil {
//			L.Push(lua.LNil)
//			L.Push(lua.LString(err.Error()))
//			return 2
//		}
//
//		dir := filepath.Dir(path)
//		if err := os.MkdirAll(dir, 0755); err != nil {
//			L.Push(lua.LNil)
//			L.Push(lua.LString(err.Error()))
//			return 2
//		}
//
//		if err := os.WriteFile(path, []byte(data), 0644); err != nil {
//			L.Push(lua.LNil)
//			L.Push(lua.LString(err.Error()))
//			return 2
//		}
//
//		L.Push(lua.LTrue)
//		return 1
//	}
//}
//
//func _scanDir(ctx *FSContext) lua.LGFunction {
//	return func(L *lua.LState) int {
//		name := L.CheckString(1)
//
//		path, err := resolvePath(ctx, name, AccessModeRead)
//		if err != nil {
//			L.Push(lua.LNil)
//			L.Push(lua.LString(err.Error()))
//			return 2
//		}
//
//		entries, err := os.ReadDir(path)
//		if err != nil {
//			L.Push(lua.LNil)
//			L.Push(lua.LString(err.Error()))
//			return 2
//		}
//
//		table := L.NewTable()
//
//		for i, entry := range entries {
//			obj := L.NewTable()
//			obj.RawSetString("name", lua.LString(entry.Name()))
//			obj.RawSetString("isDir", lua.LBool(entry.IsDir()))
//
//			table.RawSetInt(i+1, obj)
//		}
//
//		L.Push(table)
//		return 1
//	}
//}
