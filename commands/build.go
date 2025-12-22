package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sklair/caching"
	"sklair/commandRegistry"
	"sklair/discovery"
	"sklair/htmlUtilities"
	"sklair/logger"
	"sklair/sklairConfig"
	"sklair/snippets"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func init() {
	commandRegistry.Registry.Register(&commandRegistry.Command{
		Name:        "build",
		Description: "Builds a Sklair project",
		Aliases:     []string{"b"},
		Run: func(args []string) int {
			//configPath := flag.String("config", "src/sklair.json", "Path to the sklair.json config file")
			//flag.Parse()
			// TODO: fix per command flag parsing etc!
			// just extend commandregistry
			// allow per-subcommand help
			_config := "src/sklair.json"
			configPath := &_config

			config, err := sklairConfig.Load(*configPath)
			if err != nil {
				logger.Error("Could not load sklair.json : %s", err.Error())
				return 1
			}

			start := time.Now()

			configDir := filepath.Dir(*configPath)
			inputPath := filepath.Join(configDir, config.Input)
			componentsPath := filepath.Join(configDir, config.Components)
			outputPath := filepath.Join(configDir, config.Output)

			// TODO: add a function to logger which has a cool processing animation or something
			logger.Info("Indexing documents...")
			excludes := append(config.Exclude, config.Components, config.Output)
			scanned, err := discovery.DocumentDiscovery(inputPath, excludes)
			if err != nil {
				logger.Error("Could not scan documents : %s", err.Error())
				return 1
			}

			logger.Info("Indexing components...")
			components, err := discovery.ComponentDiscovery(componentsPath)
			if err != nil {
				logger.Error("Could not scan components : %s", err.Error())
				return 1
			}

			componentCache := caching.ComponentCache{
				Static:  make(map[string]*caching.Component),
				Dynamic: make(map[string]*caching.Component),
			}

			var preventFoucHead *html.Node
			var preventFoucBody *html.Node
			if config.PreventFOUC.Enabled {
				preventFoucHead, preventFoucBody, err = snippets.GetFOUCNodes(config.PreventFOUC.Colour)
				if err != nil {
					logger.Error("Could not get PreventFOUC nodes : %s", err.Error())
					return 1
				}
			}

			logger.Info("Resolving components usage and compiling...")
			for _, filePath := range scanned.HtmlFiles {
				content, err := os.ReadFile(filePath)
				if err != nil {
					logger.Error("Could not read file %s : %s", filePath, err.Error())
					return 1
				}

				//logger.Debug("File %s : %s", filePath, string(content))

				doc, err := html.Parse(bytes.NewReader(content))
				if err != nil {
					logger.Error("Could not parse file %s : %s", filePath, err.Error())
					return 1
				}

				var toReplace []*html.Node

				for node := range doc.Descendants() {
					if node.Type == html.ElementNode {
						tag := strings.ToLower(node.Data)

						if !htmlUtilities.HtmlTags[tag] {
							_, dynamicExists := componentCache.Dynamic[tag]
							_, staticExists := componentCache.Static[tag]

							if !(dynamicExists || staticExists) && tag != "lua" {
								componentSrc, exists := components[tag]
								if !exists {
									logger.Warning("Non-standard tag found in HTML and no component present : %s, assuming JS tag", tag)
									continue
								}

								logger.Info("Processing and caching tag %s...", tag)
								cached, err := caching.Cache(componentsPath, componentSrc)
								if err != nil {
									logger.Error("Could not cache component %s : %s", componentSrc, err.Error())
									return 1
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

				// TODO: LEFT OFF HERE. AFTER ALL NODES DISCOVERED ETC NEED TO REPLACE
				// TODO: in the future, hash component file contents and construct local cache in .sklair directory
				// but how would we "cache" a html.Node struct?? lol

				logger.Info("Found %d tags to replace in %s", len(toReplace), filePath)

				head := htmlUtilities.FindTag(doc, "head")

				seenComponents := make(map[string]struct{})
				seenHead := make(map[uint64]struct{})
				for _, originalTag := range toReplace {
					stcComponent, staticExists := componentCache.Static[originalTag.Data]
					dynComponent, dynamicExists := componentCache.Dynamic[originalTag.Data]

					//fmt.Println(originalTag.Data)

					if staticExists {
						parent := originalTag.Parent
						if parent != nil {
							for _, child := range stcComponent.BodyNodes {
								parent.InsertBefore(htmlUtilities.Clone(child), originalTag)
							}

							if _, seen := seenComponents[originalTag.Data]; !seen && head != nil {
								// deduplication of head nodes happens here!
								for _, child := range stcComponent.HeadNodes {
									key := htmlUtilities.WeakHashNode(child)
									//fmt.Println(key)
									if key == 0 {
										continue
									}

									if _, seen := seenHead[key]; seen {
										continue
									}
									seenHead[key] = struct{}{}
									head.AppendChild(htmlUtilities.Clone(child))
								}
							}
							seenComponents[originalTag.Data] = struct{}{}
							parent.RemoveChild(originalTag)
						}
					} else if dynamicExists {
						fmt.Println(dynComponent)
						logger.Warning("Dynamic components are not implemented yet, skipping %s...", originalTag.Data)
						continue
					} else if originalTag.Data == "lua" {
						logger.Warning("Lua components for regular input files are not implemented yet, skipping...")
						continue
					} else {
						logger.Warning("Component %s not in cache, assuming JS tag and skipping...", originalTag.Data)
						continue
					}
				}

				if config.PreventFOUC.Enabled {
					body := htmlUtilities.FindTag(doc, "body")

					if head != nil && body != nil {
						head.InsertBefore(htmlUtilities.Clone(preventFoucHead), head.FirstChild)
						body.AppendChild(htmlUtilities.Clone(preventFoucBody))
					} else {
						logger.Warning("Could not find head or body tags, skipping PreventFOUC for %s", filePath)
					}
				}

				// TODO: remove this in the future or add an option in sklair.json to disable it
				if head != nil {
					head.AppendChild(htmlUtilities.Clone(snippets.Generator))
				}

				newWriter := bytes.NewBuffer(nil)
				err = html.Render(newWriter, doc)
				if err != nil {
					logger.Error("Could not render output : %s", err.Error())
					return 1
				}

				relPath, err := filepath.Rel(inputPath, filePath)
				if err != nil {
					logger.Error("Could not get relative path : %s", err.Error())
					return 1
				}

				outPath := filepath.Join(outputPath, relPath)
				_ = os.MkdirAll(filepath.Dir(outPath), 0755)

				err = os.WriteFile(outPath, newWriter.Bytes(), 0644)
				if err != nil {
					logger.Error("Could not write output : %s", err.Error())
					return 1
				}

				logger.Info("Saved to %s", outPath)
			}

			processingEnd := time.Since(start)
			//logger.EmptyLine()
			logger.Info("Copying static files...")

			staticStart := time.Now()
			for _, filePath := range scanned.StaticFiles {
				relPath, err := filepath.Rel(inputPath, filePath)
				if err != nil {
					logger.Error("Could not get relative path for %s : %s", filePath, err.Error())
					return 1
				}

				outPath := filepath.Join(outputPath, relPath)
				_ = os.MkdirAll(filepath.Dir(outPath), 0755)

				data, err := os.ReadFile(filePath)
				if err != nil {
					logger.Error("Could not read static file %s : %s", filePath, err.Error())
					return 1
				}

				err = os.WriteFile(outPath, data, 0644)
				if err != nil {
					logger.Error("Could not write static file %s : %s", filePath, err.Error())
					return 1
				}

				logger.Info("Copied static file to %s", outPath)
			}

			//logger.EmptyLine()
			logger.Info("Compilation time (including writes) : %s", processingEnd)
			logger.Info("Static copy time : %s", time.Since(staticStart))
			logger.Info("Total processing time : %s", time.Since(start))

			return 0
		},
	})
}
