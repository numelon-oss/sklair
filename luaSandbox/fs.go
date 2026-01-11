package luaSandbox

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

type HookMode uint8

const (
	HookModePre HookMode = iota
	HookModePost
)

type FSContext struct {
	CacheDir     string
	ProjectDir   string
	TempDir      string
	GeneratedDir string
	BuiltDir     string
	Mode         HookMode
}

func openFs(opts *SandboxOptions) lua.LGFunction {
	return func(L *lua.LState) int {
		fsWithContext := make(map[string]lua.LGFunction, len(fsFuncs))
		for name, f := range fsFuncs {
			fsWithContext[name] = f(&opts.FSContext)
		}

		fsMod := L.RegisterModule("fs", fsWithContext)
		L.Push(fsMod)
		return 0
	}
}

type LFuncWithFSContext func(*FSContext) lua.LGFunction

var fsFuncs = map[string]LFuncWithFSContext{
	"read":  readFile,
	"write": writeFile,
}

type AccessMode uint8

const (
	AccessModeRead AccessMode = iota
	AccessModeWrite
)

func resolvePath(ctx *FSContext, path string, mode AccessMode) (string, error) {
	// TODO: what if theres a symlink in the parent and suddenly were exposed now?
	// yikes, but good for now because hooks are assumed trusted for now
	// TODO: later use filepath.Clean, filepath.IsAbs, filepath.Rel for full canonical validation against the above issue
	if strings.Contains(path, "..") && !(strings.Count(path, "..") == 1 && strings.HasPrefix(path, "project:../")) {
		return "", errors.New("path traversal is not allowed")
	}

	switch {
	case strings.HasPrefix(path, "cache:"):
		return filepath.Join(ctx.CacheDir, strings.TrimPrefix(path, "cache:")), nil

	case strings.HasPrefix(path, "project:"):
		if mode != AccessModeRead {
			return "", errors.New("project files are read-only")
		}
		return filepath.Join(ctx.ProjectDir, strings.TrimPrefix(path, "project:")), nil

	case strings.HasPrefix(path, "temp:"):
		return filepath.Join(ctx.TempDir, strings.TrimPrefix(path, "temp:")), nil

	case strings.HasPrefix(path, "generated:"):
		return filepath.Join(ctx.GeneratedDir, strings.TrimPrefix(path, "generated:")), nil

	case strings.HasPrefix(path, "built:"):
		if ctx.Mode != HookModePost {
			return "", errors.New("built files are only available in post-build hooks")
		}
		return filepath.Join(ctx.BuiltDir, strings.TrimPrefix(path, "built:")), nil
	}

	return "", errors.New("path must start with `cache`, `project`, `temporary`, `generated`, or `built`, followed by a colon and a relative path")
}

// readFile reads the contents of a file specified by the first argument and returns the data and a potential error.
// on success, returns the data as a string. on error, returns nil and the error message.
func readFile(ctx *FSContext) lua.LGFunction {
	return func(L *lua.LState) int {
		name := L.CheckString(1)

		path, err := resolvePath(ctx, name, AccessModeRead)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		data, err := os.ReadFile(path)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LString(data))
		return 1
	}
}

// writeFile writes the contents of the second argument to the file specified by the first argument.
// on success, returns true. on error, returns the nil and the error message.
func writeFile(ctx *FSContext) lua.LGFunction {
	return func(L *lua.LState) int {
		name := L.CheckString(1)
		data := L.CheckString(2)

		path, err := resolvePath(ctx, name, AccessModeWrite)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		if err := os.WriteFile(path, []byte(data), 0644); err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LTrue)
		return 1
	}
}

// TODO: create fs.scandir

// cache:file.txt -> .sklair/cache/file.txt
// project:file.txt (only this one allows READONLY access to one level above the project directory, using project:../file.txt)
// temporary:file.txt -> .sklair/temp/file.txt
// generated:file.txt -> .sklair/generated/file.txt -> build -> build/_sklair/generated/file.txt
// built:file.txt -> build/file.txt
