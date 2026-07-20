package building

import (
	"sklair/luaSandbox"
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
	luaRuntime := luaSandbox.NewRuntime()
	defer luaRuntime.Close()
	documentQueue := newDocumentQueue(inputs.layouts)

	preHookStart := time.Now()
	if err := runPreHooks(inputs, luaRuntime, documentQueue); err != nil {
		return err
	}
	preHookEnd := time.Since(preHookStart)

	plan, err := planBuild(inputs, documentQueue)
	if err != nil {
		return err
	}
	if generated := plan.generatedCount(); generated > 0 {
		logger.Info("Planned %d generated documents", generated)
	}

	compilationStart := time.Now()
	logger.Info("Preparing document definitions...")
	definitions, err := prepareDefinitions(inputs, plan, luaRuntime)
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
	if err := runPostHooks(inputs, luaRuntime); err != nil {
		return err
	}
	postHookEnd := time.Since(postHookStart)

	//logger.EmptyLine()
	logger.Info("Compilation (including writes) of %d documents (%d generated) : %s", len(plan.documents), plan.generatedCount(), processingEnd)
	logger.Info("Static copy of %d files : %s", len(plan.staticFiles), staticEnd)
	if inputs.hooks != nil {
		logger.Info("Run time of %d pre-build hooks : %s", len(inputs.hooks.PreBuild), preHookEnd)
		logger.Info("Run time of %d post-build hooks : %s", len(inputs.hooks.PostBuild), postHookEnd)
	}
	logger.Info("Time since start : %s", time.Since(start))

	return nil
}
