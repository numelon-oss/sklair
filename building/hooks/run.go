package hooks

import (
	"fmt"
	"path/filepath"
	"sklair/luaSandbox"
)

func RunHooks(hooksDir string, hooks []string, ctx *luaSandbox.FSContext) error {
	which := "pre"
	if ctx.Mode == luaSandbox.HookModePost {
		which = "post"
	}

	hookDir := filepath.Join(hooksDir, which)

	exitChannel := make(chan int)

	for _, hookFilename := range hooks {
		sourcePath := filepath.Join(hookDir, hookFilename)

		exitChannel = make(chan int, 1)
		done := make(chan error, 1)

		L := luaSandbox.NewSandbox(luaSandbox.SandboxOptions{
			ExitChannel: exitChannel,
			FSContext:   *ctx,
		})

		// lua must run asynchronously, otherwise we cannot track when the exit channel was used with os.exit()
		go func() {
			done <- L.DoFile(sourcePath)
		}()

		select {
		case code := <-exitChannel:
			// lua called os.exit()
			L.Close()

			switch code {
			case 0:
				return nil
			case 1:
				return fmt.Errorf("hook %s exited with failure", hookFilename)
			default:
				return fmt.Errorf("hook %s exited with code %d", hookFilename, code)
			}

		case err := <-done:
			L.Close()
			if err != nil {
				return fmt.Errorf("hook %s failed\n%s", hookFilename, err.Error())
			}
		}
	}

	return nil
}
