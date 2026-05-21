package building

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sklair/building/hooks"
	"sklair/building/priorities"
	"sklair/building/resources"
	"sklair/caching"
	"sklair/devserver"
	"sklair/discovery"
	"sklair/htmlUtilities"
	"sklair/luaSandbox"
	sklairRuntime "sklair/runtime"
	"sklair/sklairConfig"
	"sklair/snippets"
	"sklair/util"
	"strings"
	"time"

	"github.com/numelon-oss/go-logger"

	"golang.org/x/net/html"
)

func Build(config *sklairConfig.ProjectConfig, configDir string, outputDirOverride string) error {
	start := time.Now()

	inputDir := filepath.Join(configDir, config.Input)
	componentsDir := filepath.Join(configDir, config.Components)
	hooksPath := ""
	if config.Hooks != nil && config.Hooks.Enabled {
		hooksPath = config.Hooks.Path
	}
	hooksDir := filepath.Join(configDir, hooksPath)

	outputDir := outputDirOverride
	if outputDirOverride == "" {
		outputDir = filepath.Join(configDir, config.Output)
	}

	sklairDir := filepath.Join(configDir, ".sklair")
	cacheDir := filepath.Join(sklairDir, "cache")
	tempDir := filepath.Join(sklairDir, "temp")
	generatedDir := filepath.Join(sklairDir, "generated")

	componentsRel, err := filepath.Rel(inputDir, componentsDir)
	hooksRel, err := filepath.Rel(inputDir, hooksDir)
	if err != nil {
		return errors.New("could not get relative path for components or hooks : " + err.Error())
	}
	excludes := append(config.Exclude, componentsRel, hooksRel)

	if outputDirOverride == "" {
		outputRel, err := filepath.Rel(inputDir, outputDir)
		if err != nil {
			return errors.New("could not get relative path for output : " + err.Error())
		}
		excludes = append(excludes, outputRel)
	}

	err = os.RemoveAll(outputDir)
	if err != nil {
		return fmt.Errorf("could not remove output directory %s : %s", outputDir, err.Error())
	}
	err = os.RemoveAll(tempDir)
	if err != nil {
		return fmt.Errorf("could not remove Sklair's temp directory %s : %s", outputDir, err.Error())
	}
	err = os.RemoveAll(generatedDir)
	if err != nil {
		return fmt.Errorf("could not remove Sklair's generated directory %s : %s", outputDir, err.Error())
	}

	logger.Info("Indexing documents...")
	scanned, err := discovery.DiscoverDocuments(inputDir, excludes, config.ExcludeCompile)
	if err != nil {
		return errors.New("could not scan documents : " + err.Error())
	}

	logger.Info("Indexing components...")
	components, err := discovery.DiscoverComponents(componentsDir)
	if err != nil {
		return errors.New("could not scan components : " + err.Error())
	}

	// templates are literally just components but are treated specially
	// i.e. they cannot be inserted more than once
	templates := make(map[string]struct{})
	for _, name := range config.Templates {
		tag := strings.ToLower(strings.TrimSpace(name))
		if tag == "" {
			return errors.New("runtime template component name cannot be empty")
		}
		if _, exists := components[tag]; !exists {
			return fmt.Errorf("runtime template component %q does not exist", name)
		}
		templates[tag] = struct{}{}
	}

	// TODO: hooks are really messy here, especially the allHooks variable (potential nil reference later)
	// so later rewrite some of it to be more readable and less error prone
	// perhaps just abstract the entire hooks system into a function dedicated for this build step only?
	// also rename the luaSandbox package to "hooks" because it makes more sense (or maybe dont)
	preHookStart := time.Now()
	var allHooks *discovery.Hookset
	if config.Hooks != nil && config.Hooks.Enabled {
		logger.Info("Indexing hooks...")
		allHooks, err = discovery.DiscoverHooks(hooksDir)
		if err != nil {
			return errors.New("could not scan hooks : " + err.Error())
		}

		logger.Info("Running pre-build hooks...")
		err = hooks.RunHooks(hooksDir, allHooks.PreBuild, &luaSandbox.FSContext{
			CacheDir:     cacheDir,
			ProjectDir:   inputDir,
			TempDir:      tempDir,
			GeneratedDir: generatedDir,
			BuiltDir:     outputDir,
			Mode:         luaSandbox.HookModePre,
		})
		if err != nil {
			return fmt.Errorf("could not run pre-build hooks : %s", err.Error())
		}
	}
	preHookEnd := time.Since(preHookStart)

	componentCache := caching.ComponentCache{
		Static:  make(map[string]*caching.Component),
		Dynamic: make(map[string]*caching.Component),
	}
	usedComponentFolders := make(map[string]discovery.ComponentSource)
	runtimeUsed := false

	var preventFoucHead *html.Node
	var preventFoucBody *html.Node
	if config.PreventFOUC != nil && config.PreventFOUC.Enabled {
		preventFoucHead, preventFoucBody, err = snippets.GetFOUCNodes(config.PreventFOUC.Colour)
		if err != nil {
			return errors.New("could not get PreventFOUC nodes : " + err.Error())
		}
	}

	compilationStart := time.Now()

	logger.Info("Resolving components usage and compiling...")
	for _, filePath := range scanned.HtmlFiles {
		doc, err := htmlUtilities.ParseFile(filePath)
		if err != nil {
			return fmt.Errorf("could not parse file %s : %s", filePath, err.Error())
		}

		var toReplace []*html.Node

		// lazy components below

		for node := range doc.Descendants() {
			if node.Type == html.ElementNode {
				tag := strings.ToLower(node.Data)

				if !htmlUtilities.HtmlTags[tag] {
					_, dynamicExists := componentCache.Dynamic[tag]
					_, staticExists := componentCache.Static[tag]

					if !(dynamicExists || staticExists) && (!(tag == "lua" || tag == "opengraph")) {
						componentSrc, exists := components[tag]
						if !exists {
							logger.Warning("Non-standard tag found in HTML and no component present : %s; assuming Autonomous Custom Element", tag)
							continue
						}

						logger.Info("Processing and caching tag %s...", tag)
						cached, err := caching.MakeCache(componentsDir, componentSrc.Entry())
						if err != nil {
							return fmt.Errorf("could not cache component %s : %s", componentSrc.Entry(), err.Error())
						}
						if componentSrc.IsFolder {
							assetOutputDir := "/_sklair/components/" + tag
							if err := resources.RewriteURLs(cached.HeadNodes, assetOutputDir); err != nil {
								return fmt.Errorf("could not rewrite component %s head resource URLs : %s", componentSrc.Entry(), err.Error())
							}
							if err := resources.RewriteURLs(cached.BodyNodes, assetOutputDir); err != nil {
								return fmt.Errorf("could not rewrite component %s body resource URLs : %s", componentSrc.Entry(), err.Error())
							}
						}

						if cached.Dynamic {
							componentCache.Dynamic[tag] = cached
						} else {
							componentCache.Static[tag] = cached
						}
					}

					toReplace = append(toReplace, node)
				}
			}
		}

		// TODO: in the future, hash component file contents and construct local cache in .sklair directory
		// but how would we "cache" a html.Node struct?? lol

		logger.Info("Found %d tags to replace in %s", len(toReplace), filePath)

		head := htmlUtilities.FindTag(doc, "head")
		body := htmlUtilities.FindTag(doc, "body")
		if head == nil || body == nil {
			return fmt.Errorf("could not find head or body tags in %s, how does that even happen", filePath)
		}

		// usedComponents ensures that each component contributes its <head> nodes at most ONCE per document,
		// even if the component appears multiple times in the source document
		usedComponents := make(map[string]struct{})
		// seenHead, on the other hand, is used for actual deduplication
		registeredRuntimeTemplates := make(map[string]struct{})
		for _, originalTag := range toReplace {
			tag := strings.ToLower(originalTag.Data)
			stcComponent, staticExists := componentCache.Static[tag]
			dynComponent, dynamicExists := componentCache.Dynamic[tag]

			parent := originalTag.Parent
			if parent == nil {
				return fmt.Errorf("somehow the parent does not exist for %s. (memory corruption???)", originalTag.Data)
			}

			//fmt.Println(originalTag.Data)

			if staticExists || dynamicExists {
				if err := checkNoBody(originalTag); err != nil {
					return fmt.Errorf("invalid use of component %s in %s : %s", originalTag.Data, filePath, err.Error())
				}
			}

			// TODO: the logic for static and dynamic components will likely be very similar
			// in the future, simply combine both branches,
			// but for dynamic components just have a simple processing stage.
			// after that its treated as a static component would be
			if staticExists {
				// this check ensures that each component contributes its <head> nodes at most ONCE per document
				if _, seen := usedComponents[tag]; !seen {
					htmlUtilities.AppendNodes(head, stcComponent.HeadNodes)
				}
				usedComponents[tag] = struct{}{}
				if source := components[tag]; source.IsFolder {
					usedComponentFolders[tag] = source
				}

				if _, isRuntimeTemplate := templates[tag]; isRuntimeTemplate {
					if _, registered := registeredRuntimeTemplates[tag]; registered {
						return fmt.Errorf("runtime template component %s is registered more than once in %s", originalTag.Data, filePath)
					}

					template := &html.Node{
						Type: html.ElementNode,
						Data: "template",
						Attr: []html.Attribute{{Key: "id", Val: "sklair-template-" + tag}},
					}
					for _, node := range stcComponent.BodyNodes {
						template.AppendChild(htmlUtilities.Clone(node))
					}
					parent.InsertBefore(template, originalTag)
					registeredRuntimeTemplates[tag] = struct{}{}
					runtimeUsed = true
				} else {
					htmlUtilities.InsertNodesBefore(originalTag, stcComponent.BodyNodes)
				}

				parent.RemoveChild(originalTag)
			} else if dynamicExists {
				if _, isRuntimeTemplate := templates[tag]; isRuntimeTemplate {
					return fmt.Errorf("runtime template component %s is dynamic, but dynamic components are not implemented yet", originalTag.Data)
				}
				fmt.Println(dynComponent)
				logger.Warning("Dynamic components are not implemented yet, skipping %s...", originalTag.Data)
				continue
			} else if originalTag.Data == "lua" {
				// TODO: prints from lua will be appended to a buffer
				// then this buffer will be parsed by html
				// then this will be inserted into document
				// TODO: or should we actually instead expose a library eg `sklair` and we can do `sklair.put()`? thats probably cleaner
				// and also easier to implement
				logger.Warning("Lua components for regular input files are not implemented yet, skipping...")
				continue
			} else if originalTag.Data == "opengraph" {
				for _, child := range snippets.OpenGraph(originalTag) {
					head.AppendChild(child)
				}
				parent.RemoveChild(originalTag)
			} else {
				logger.Warning("Component %s not in cache; assuming unregistered custom element and skipping...", originalTag.Data)
				continue
			}
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

		relPath, err := filepath.Rel(inputDir, filePath)
		if err != nil {
			return fmt.Errorf("could not get relative path for %s : %s", filePath, err.Error())
		}

		outPath := filepath.Join(outputDir, relPath)
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
		if err := copyComponentFolders(componentsDir, outputDir, usedComponentFolders); err != nil {
			return err
		}
	}

	if runtimeUsed {
		path := filepath.Join(outputDir, filepath.FromSlash(sklairRuntime.OutputPath))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("could not create Sklair runtime directory : %s", err.Error())
		}
		if err := os.WriteFile(path, []byte(sklairRuntime.Module), 0644); err != nil {
			return fmt.Errorf("could not write Sklair runtime module : %s", err.Error())
		}
	}

	if outputDirOverride != "" {
		err = os.MkdirAll(filepath.Join(outputDir, "_sklair"), 0755)
		if err != nil {
			return fmt.Errorf("could not create sklair dev server directory : %s", err.Error())
		}

		err := os.WriteFile(filepath.Join(outputDir, devserver.WSDevScriptPath), []byte(devserver.WSDevScript), 0644)
		if err != nil {
			return fmt.Errorf("could not write sklair dev server websocket js file : %s", err.Error())
		}
	}

	//logger.EmptyLine()
	logger.Info("Copying static files...")

	staticStart := time.Now()
	for _, filePath := range scanned.StaticFiles {
		relPath, err := filepath.Rel(inputDir, filePath)
		if err != nil {
			return fmt.Errorf("could not get relative path for %s : %s", filePath, err.Error())
		}

		outPath := filepath.Join(outputDir, relPath)
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
	if allHooks != nil {
		buildSklairDir := filepath.Join(outputDir, "_sklair") // TODO: the _sklair directory in output is not unique to hooks, they will be used for more things in the future

		isEmpty, err := util.IsDirEmpty(generatedDir)
		if err != nil {
			exist := os.IsExist(err)
			if exist {
				return fmt.Errorf("could not check if generated directory is empty : %s", err.Error())
			} else {
				isEmpty = true
			}
		}
		if !isEmpty {
			err = util.CopyDir(generatedDir, buildSklairDir)
			if err != nil {
				return fmt.Errorf("could not copy generated files to Sklair's namespace : %s", err.Error())
			}
		}

		logger.Info("Running post-build hooks...")
		err = hooks.RunHooks(hooksDir, allHooks.PostBuild, &luaSandbox.FSContext{
			CacheDir:     cacheDir,
			ProjectDir:   inputDir,
			TempDir:      tempDir,
			GeneratedDir: buildSklairDir,
			BuiltDir:     outputDir,
			Mode:         luaSandbox.HookModePost,
		})
		if err != nil {
			return fmt.Errorf("could not run post-build hooks : %s", err.Error())
		}
	}
	postHookEnd := time.Since(postHookStart)

	//logger.EmptyLine()
	logger.Info("Compilation (including writes) of %d files : %s", len(scanned.HtmlFiles), processingEnd)
	logger.Info("Static copy of %d files : %s", len(scanned.StaticFiles), staticEnd)
	if allHooks != nil {
		logger.Info("Run time of %d pre-build hooks : %s", len(allHooks.PreBuild), preHookEnd)
		logger.Info("Run time of %d post-build hooks : %s", len(allHooks.PostBuild), postHookEnd)
	}
	logger.Info("Time since start : %s", time.Since(start))

	return nil
}
