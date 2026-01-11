package luaSandbox

import lua "github.com/yuin/gopher-lua"

type SandboxOptions struct {
	ExitChannel chan int
	FSContext   FSContext
}

// NewSandbox creates a new Lua state with default Lua libraries opened but cleaned or modified to create a sandboxed environment.
// It also loads our own custom libraries.
func NewSandbox(options SandboxOptions) *lua.LState {
	L := lua.NewState(lua.Options{
		RegistrySize:        128,
		RegistryMaxSize:     512,
		SkipOpenLibs:        true,
		IncludeGoStackTrace: false,
		//MinimizeStackMemory: false,
	})
	
	OpenSandboxedDefault(L, &options)
	OpenSandboxedCustom(L, &options)

	return L
}
