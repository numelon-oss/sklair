package building

import (
	"fmt"
	"sklair/building/hooks"
	"sklair/luaSandbox"

	"github.com/numelon-oss/go-logger"
)

func runPreHooks(inputs *buildInputs) error {
	if inputs.hooks == nil {
		return nil
	}

	logger.Info("Running pre-build hooks...")
	if err := hooks.RunHooks(inputs.paths.hooks, inputs.hooks.PreBuild, &luaSandbox.FSContext{
		CacheDir:     inputs.paths.cache,
		ProjectDir:   inputs.paths.input,
		TempDir:      inputs.paths.temp,
		GeneratedDir: inputs.paths.generated,
		BuiltDir:     inputs.paths.output,
		Mode:         luaSandbox.HookModePre,
	}); err != nil {
		return fmt.Errorf("could not run pre-build hooks : %s", err.Error())
	}

	return nil
}
