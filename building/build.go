package building

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sklair/building/hooks"
	"sklair/building/priorities"
	"sklair/devserver"
	"sklair/discovery"
	"sklair/htmlUtilities"
	"sklair/luaSandbox"
	sklairRuntime "sklair/runtime"
	"sklair/sklairConfig"
	"sklair/snippets"
	"sklair/util"
	"sort"
	"time"

	"github.com/numelon-oss/go-logger"

	"golang.org/x/net/html"
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
	usedComponentFolders := make(map[string]discovery.ComponentSource)

	logger.Info("Resolving components usage and compiling...")
	documents := make([]*documentState, 0, len(definitions.documents))
	for _, definition := range definitions.documents {
		document, err := compileDocument(definition, componentResolver)
		if err != nil {
			return err
		}
		documents = append(documents, document)
		for name, source := range document.componentFolders {
			usedComponentFolders[name] = source
		}
	}

	var preventFoucHead *html.Node
	var preventFoucBody *html.Node
	if config.PreventFOUC != nil && config.PreventFOUC.Enabled {
		preventFoucHead, preventFoucBody, err = snippets.GetFOUCNodes(config.PreventFOUC.Colour)
		if err != nil {
			return errors.New("could not get PreventFOUC nodes : " + err.Error())
		}
	}

	err = os.RemoveAll(inputs.paths.output)
	if err != nil {
		return fmt.Errorf("could not remove output directory %s : %s", inputs.paths.output, err.Error())
	}

	runtimeUsed := false
	for _, document := range documents {
		filePath := document.source
		doc := document.root
		head := htmlUtilities.FindTag(doc, "head")
		body := htmlUtilities.FindTag(doc, "body")
		requiredRuntimeTemplates := document.templates

		if len(requiredRuntimeTemplates) > 0 {
			templateNames := make([]string, 0, len(requiredRuntimeTemplates))
			for name := range requiredRuntimeTemplates {
				templateNames = append(templateNames, name)
			}
			sort.Strings(templateNames)

			registry := &html.Node{
				Type: html.ElementNode,
				Data: "div",
				Attr: []html.Attribute{
					{Key: "id", Val: "_sklair-runtime-templates"},
					{Key: "hidden"},
				},
			}
			for _, name := range templateNames {
				component := requiredRuntimeTemplates[name]

				template := &html.Node{
					Type: html.ElementNode,
					Data: "template",
					Attr: []html.Attribute{{Key: "id", Val: "sklair-template-" + name}},
				}
				for _, node := range component.BodyNodes {
					template.AppendChild(htmlUtilities.Clone(node))
				}
				registry.AppendChild(template)
			}
			body.AppendChild(registry)
			runtimeUsed = true
		}

		// --------------------------------------------------
		// resource hints
		// --------------------------------------------------

		// TODO: if google found in link rel for google fonts, then add preconnect for fonts.gstatic.com
		// basically for known preconnects

		// cap preconnect to 6 origins
		// warn if more than 6 and consider self hosting some assets
		// ensure google fonts is cross origin
		// todo image srcset
		// https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Attributes/rel/preconnect
		//origins := make(map[string]int)
		//if config.ResourceHints != nil && config.ResourceHints.Enabled {
		//	for node := range doc.Descendants() {
		//		if node.Type == html.ElementNode {
		//
		//		}
		//	}
		//}

		// --------------------------------------------------
		// head segmentation and optimisation
		// --------------------------------------------------
		segmentedHead, err := SegmentHead(head)
		if err != nil {
			return fmt.Errorf("could not segment <head> in %s : %s", filePath, err.Error())
		}

		if config.PreventFOUC != nil && config.PreventFOUC.Enabled {
			segmentedHead = append(segmentedHead, &HeadSegment{
				Nodes:             []*html.Node{htmlUtilities.Clone(preventFoucHead)},
				TreatAsTag:        priorities.PreventFOUC,
				IsOrderingBarrier: false,
			})

			body.AppendChild(htmlUtilities.Clone(preventFoucBody))
		}

		// TODO: remove this (generator) in the future or add an option in sklair.json to disable it
		segmentedHead = append(segmentedHead, &HeadSegment{
			Nodes:             []*html.Node{htmlUtilities.Clone(snippets.Generator)},
			TreatAsTag:        priorities.Generator,
			IsOrderingBarrier: false,
		})

		if outputDirOverride != "" {
			// sklair dev server refresh with websocket
			segmentedHead = append(segmentedHead, &HeadSegment{
				Nodes: []*html.Node{
					htmlUtilities.Clone(devserver.WSScriptNode),
				},
				TreatAsTag:        priorities.Script,
				IsOrderingBarrier: false,
			})
		}

		segmentedHead = OptimiseHead(segmentedHead)

		// put the segmented head back into the document head
		htmlUtilities.RemoveAllChildren(head)
		for _, seg := range segmentedHead {
			for _, node := range seg.Nodes {
				head.AppendChild(node) // no need to clone because everything was either already cloned before, OR is already from the same document
			}
		}

		newWriter := bytes.NewBuffer(nil)
		err = html.Render(newWriter, doc)
		if err != nil {
			return fmt.Errorf("could not render output for %s : %s", filePath, err.Error())
		}

		outPath := document.output
		err = os.MkdirAll(filepath.Dir(outPath), 0755)
		if err != nil {
			return fmt.Errorf("could not create output directory for %s : %s", filePath, err.Error())
		}

		err = os.WriteFile(outPath, newWriter.Bytes(), 0644)
		if err != nil {
			return fmt.Errorf("could not write output for %s : %s", filePath, err.Error())
		}

		logger.Info("Saved to %s", outPath)
	}

	processingEnd := time.Since(compilationStart)

	if len(usedComponentFolders) > 0 {
		logger.Info("Copying component assets...")
		if err := copyComponentFolders(inputs.paths.components, inputs.paths.output, usedComponentFolders); err != nil {
			return err
		}
	}

	if runtimeUsed {
		path := filepath.Join(inputs.paths.output, filepath.FromSlash(sklairRuntime.OutputPath))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("could not create Sklair runtime directory : %s", err.Error())
		}
		if err := os.WriteFile(path, []byte(sklairRuntime.Module), 0644); err != nil {
			return fmt.Errorf("could not write Sklair runtime module : %s", err.Error())
		}
	}

	if outputDirOverride != "" {
		err = os.MkdirAll(filepath.Join(inputs.paths.output, "_sklair"), 0755)
		if err != nil {
			return fmt.Errorf("could not create sklair dev server directory : %s", err.Error())
		}

		err := os.WriteFile(filepath.Join(inputs.paths.output, devserver.WSDevScriptPath), []byte(devserver.WSDevScript), 0644)
		if err != nil {
			return fmt.Errorf("could not write sklair dev server websocket js file : %s", err.Error())
		}
	}

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
