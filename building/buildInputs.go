package building

import (
	"errors"
	"fmt"
	"path/filepath"
	"sklair/discovery"
	"sklair/sklairConfig"
	"strings"

	"github.com/numelon-oss/go-logger"
)

type buildPaths struct {
	input      string
	components string
	hooks      string
	output     string
	cache      string
	temp       string
	generated  string
}

type buildInputs struct {
	paths      buildPaths
	documents  *discovery.DocumentLists
	components map[string]discovery.ComponentSource
	hooks      *discovery.Hookset
	templates  map[string]struct{}
}

func discoverBuild(config *sklairConfig.ProjectConfig, configDir string, outputDirOverride string) (*buildInputs, error) {
	paths := resolveBuildPaths(config, configDir, outputDirOverride)

	excludes, err := buildExcludes(config, paths, outputDirOverride != "")
	if err != nil {
		return nil, err
	}

	logger.Info("Indexing documents...")
	documents, err := discovery.DiscoverDocuments(paths.input, excludes, config.ExcludeCompile)
	if err != nil {
		return nil, errors.New("could not scan documents : " + err.Error())
	}

	logger.Info("Indexing components...")
	components, err := discovery.DiscoverComponents(paths.components)
	if err != nil {
		return nil, errors.New("could not scan components : " + err.Error())
	}

	templates, err := classifyTemplates(config.Templates, components)
	if err != nil {
		return nil, err
	}

	var discoveredHooks *discovery.Hookset
	if config.Hooks != nil && config.Hooks.Enabled {
		logger.Info("Indexing hooks...")
		discoveredHooks, err = discovery.DiscoverHooks(paths.hooks)
		if err != nil {
			return nil, errors.New("could not scan hooks : " + err.Error())
		}
	}

	return &buildInputs{
		paths:      paths,
		documents:  documents,
		components: components,
		hooks:      discoveredHooks,
		templates:  templates,
	}, nil
}

func resolveBuildPaths(config *sklairConfig.ProjectConfig, configDir string, outputDirOverride string) buildPaths {
	inputDir := filepath.Join(configDir, config.Input)
	componentsDir := filepath.Join(configDir, config.Components)

	hooksPath := ""
	if config.Hooks != nil && config.Hooks.Enabled {
		hooksPath = config.Hooks.Path
	}
	hooksDir := filepath.Join(configDir, hooksPath)

	outputDir := outputDirOverride
	if outputDir == "" {
		outputDir = filepath.Join(configDir, config.Output)
	}

	sklairDir := filepath.Join(configDir, ".sklair")

	return buildPaths{
		input:      inputDir,
		components: componentsDir,
		hooks:      hooksDir,
		output:     outputDir,
		cache:      filepath.Join(sklairDir, "cache"),
		temp:       filepath.Join(sklairDir, "temp"),
		generated:  filepath.Join(sklairDir, "generated"),
	}
}

func buildExcludes(config *sklairConfig.ProjectConfig, paths buildPaths, outputOverridden bool) ([]string, error) {
	componentsRel, err := filepath.Rel(paths.input, paths.components)
	if err != nil {
		return nil, errors.New("could not get relative path for components : " + err.Error())
	}

	hooksRel, err := filepath.Rel(paths.input, paths.hooks)
	if err != nil {
		return nil, errors.New("could not get relative path for hooks : " + err.Error())
	}

	excludes := append([]string{}, config.Exclude...)
	excludes = append(excludes, componentsRel, hooksRel)

	if !outputOverridden {
		outputRel, err := filepath.Rel(paths.input, paths.output)
		if err != nil {
			return nil, errors.New("could not get relative path for output : " + err.Error())
		}
		excludes = append(excludes, outputRel)
	}

	return excludes, nil
}

func classifyTemplates(names []string, components map[string]discovery.ComponentSource) (map[string]struct{}, error) {
	templates := make(map[string]struct{}, len(names))

	for _, name := range names {
		tag := strings.ToLower(strings.TrimSpace(name))
		if tag == "" {
			return nil, errors.New("runtime template component name cannot be empty")
		}
		if _, exists := components[tag]; !exists {
			return nil, fmt.Errorf("runtime template component %q does not exist", name)
		}
		templates[tag] = struct{}{}
	}

	return templates, nil
}
