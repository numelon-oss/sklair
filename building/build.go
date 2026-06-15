package building

import (
	"fmt"
	"os"
	"path/filepath"
	"sklair/building/hooks"
	"sklair/luaSandbox"
	"sklair/sklairConfig"
	"sklair/util"
	"time"

	"github.com/numelon-oss/go-logger"
)

func Build(config *sklairConfig.ProjectConfig, configDir string, outputDirOverride string) error {
	start := time.Now()

	inputs, err := discoverBuild(config, configDir, outputDirOverride)
	if err != nil {
		return err
	}

	err = os.RemoveAll(inputs.paths.temp)
	if err != nil {
		return fmt.Errorf("could not remove Sklair's temp directory %s : %s", inputs.paths.temp, err.Error())
	}
	err = os.RemoveAll(inputs.paths.generated)
	if err != nil {
		return fmt.Errorf("could not remove Sklair's generated directory %s : %s", inputs.paths.generated, err.Error())
	}

	preHookStart := time.Now()
	if err := runPreHooks(inputs); err != nil {
		return err
	}
	preHookEnd := time.Since(preHookStart)

	plan, err := planBuild(inputs)
	if err != nil {
		return err
	}

	compilationStart := time.Now()
	logger.Info("Preparing document definitions...")
	definitions, err := prepareDefinitions(inputs, plan)
	if err != nil {
		return err
	}

	componentResolver := newComponentResolver(definitions.components, inputs.templates)

	logger.Info("Resolving components usage and compiling...")
	documents := make([]*documentState, 0, len(definitions.documents))
	for _, definition := range definitions.documents {
		document, err := compileDocument(definition, componentResolver)
		if err != nil {
			return err
		}
		documents = append(documents, document)
	}

	logger.Info("Finalising documents...")
	output, err := finaliseDocuments(documents, config, outputDirOverride != "")
	if err != nil {
		return err
	}

	err = os.RemoveAll(inputs.paths.output)
	if err != nil {
		return fmt.Errorf("could not remove output directory %s : %s", inputs.paths.output, err.Error())
	}

	if err := writeOutput(output, inputs.paths); err != nil {
		return err
	}
	processingEnd := time.Since(compilationStart)

	//logger.EmptyLine()
	logger.Info("Copying static files...")

	staticStart := time.Now()
	for _, planned := range plan.staticFiles {
		filePath := planned.source
		outPath := planned.output
		err = os.MkdirAll(filepath.Dir(outPath), 0755)
		if err != nil {
			return fmt.Errorf("could not create output directory for %s : %s", filePath, err.Error())
		}

		err = util.CopyFile(filePath, outPath, 0644)
		if err != nil {
			return fmt.Errorf("could not copy static file %s : %s", filePath, err.Error())
		}

		logger.Info("Copied static file to %s", outPath)
	}

	staticEnd := time.Since(staticStart)

	postHookStart := time.Now()
	if inputs.hooks != nil {
		buildSklairDir := filepath.Join(inputs.paths.output, "_sklair") // TODO: the _sklair directory in output is not unique to hooks, they will be used for more things in the future

		isEmpty, err := util.IsDirEmpty(inputs.paths.generated)
		if err != nil {
			exist := os.IsExist(err)
			if exist {
				return fmt.Errorf("could not check if generated directory is empty : %s", err.Error())
			} else {
				isEmpty = true
			}
		}
		if !isEmpty {
			err = util.CopyDir(inputs.paths.generated, buildSklairDir)
			if err != nil {
				return fmt.Errorf("could not copy generated files to Sklair's namespace : %s", err.Error())
			}
		}

		logger.Info("Running post-build hooks...")
		err = hooks.RunHooks(inputs.paths.hooks, inputs.hooks.PostBuild, &luaSandbox.FSContext{
			CacheDir:     inputs.paths.cache,
			ProjectDir:   inputs.paths.input,
			TempDir:      inputs.paths.temp,
			GeneratedDir: buildSklairDir,
			BuiltDir:     inputs.paths.output,
			Mode:         luaSandbox.HookModePost,
		})
		if err != nil {
			return fmt.Errorf("could not run post-build hooks : %s", err.Error())
		}
	}
	postHookEnd := time.Since(postHookStart)

	//logger.EmptyLine()
	logger.Info("Compilation (including writes) of %d files : %s", len(plan.documents), processingEnd)
	logger.Info("Static copy of %d files : %s", len(plan.staticFiles), staticEnd)
	if inputs.hooks != nil {
		logger.Info("Run time of %d pre-build hooks : %s", len(inputs.hooks.PreBuild), preHookEnd)
		logger.Info("Run time of %d post-build hooks : %s", len(inputs.hooks.PostBuild), postHookEnd)
	}
	logger.Info("Time since start : %s", time.Since(start))

	return nil
}
