package building

import (
	"sklair/sklairConfig"
	"time"

	"github.com/numelon-oss/go-logger"
)

func Build(config *sklairConfig.ProjectConfig, configDir string, outputDirOverride string) error {
	start := time.Now()

	inputs, err := discoverBuild(config, configDir, outputDirOverride)
	if err != nil {
		return err
	}

	if err := resetHookWorkspace(inputs.paths); err != nil {
		return err
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

	logger.Info("Resolving components usage and compiling...")
	documents, err := compileDocuments(definitions, inputs.templates)
	if err != nil {
		return err
	}

	logger.Info("Finalising documents...")
	output, err := finaliseDocuments(documents, config, outputDirOverride != "")
	if err != nil {
		return err
	}

	if err := replaceOutput(output, inputs.paths); err != nil {
		return err
	}
	processingEnd := time.Since(compilationStart)

	staticStart := time.Now()
	if err := copyStatic(plan.staticFiles); err != nil {
		return err
	}
	staticEnd := time.Since(staticStart)

	postHookStart := time.Now()
	if err := runPostHooks(inputs); err != nil {
		return err
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
