package building

import (
	"fmt"
	"os"
	"path/filepath"
	"sklair/building/hooks"
	"sklair/luaSandbox"
	"sklair/util"

	"github.com/numelon-oss/go-logger"
)

func resetHookWorkspace(paths buildPaths) error {
	if err := os.RemoveAll(paths.temp); err != nil {
		return fmt.Errorf("could not remove Sklair's temp directory %s : %s", paths.temp, err.Error())
	}
	if err := os.RemoveAll(paths.generated); err != nil {
		return fmt.Errorf("could not remove Sklair's generated directory %s : %s", paths.generated, err.Error())
	}

	return nil
}

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

func runPostHooks(inputs *buildInputs) error {
	if inputs.hooks == nil {
		return nil
	}

	buildSklairDir := filepath.Join(inputs.paths.output, "_sklair") // TODO: the _sklair directory in output is not unique to hooks, they will be used for more things in the future

	isEmpty, err := util.IsDirEmpty(inputs.paths.generated)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("could not check if generated directory is empty : %s", err.Error())
		}
		isEmpty = true
	}
	if !isEmpty {
		if err := util.CopyDir(inputs.paths.generated, buildSklairDir); err != nil {
			return fmt.Errorf("could not copy generated files to Sklair's namespace : %s", err.Error())
		}
	}

	logger.Info("Running post-build hooks...")
	if err := hooks.RunHooks(inputs.paths.hooks, inputs.hooks.PostBuild, &luaSandbox.FSContext{
		CacheDir:     inputs.paths.cache,
		ProjectDir:   inputs.paths.input,
		TempDir:      inputs.paths.temp,
		GeneratedDir: buildSklairDir,
		BuiltDir:     inputs.paths.output,
		Mode:         luaSandbox.HookModePost,
	}); err != nil {
		return fmt.Errorf("could not run post-build hooks : %s", err.Error())
	}

	return nil
}
