package hooks

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sklair/luaSandbox"
)

func RunHooks(runtime *luaSandbox.Runtime, hooksDir string, hooks []string, ctx *luaSandbox.FSContext) error {
	which := "pre"
	if ctx.Mode == luaSandbox.HookModePost {
		which = "post"
	}

	hookDir := filepath.Join(hooksDir, which)

	for _, hookFilename := range hooks {
		sourcePath := filepath.Join(hookDir, hookFilename)
		scope := runtime.NewScope(luaSandbox.SandboxOptions{
			FSContext: *ctx,
			Profile:   luaSandbox.HookSandbox,
		})
		source, err := os.ReadFile(sourcePath)
		if err != nil {
			return fmt.Errorf("hook %s failed\n%s", hookFilename, err.Error())
		}
		result, runErr := scope.Run(bytes.NewReader(source), sourcePath)
		if runErr != nil {
			return fmt.Errorf("hook %s failed\n%s", hookFilename, runErr.Error())
		}
		if result.Exited {
			switch result.ExitCode {
			case 0:
				continue
			case 1:
				return fmt.Errorf("hook %s exited with failure", hookFilename)
			default:
				return fmt.Errorf("hook %s exited with code %d", hookFilename, result.ExitCode)
			}
		}
	}

	return nil
}
