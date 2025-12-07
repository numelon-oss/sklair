package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sklair/caching"
	"sklair/discovery"
	"sklair/htmlUtilities"
	"sklair/logger"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const ComponentsDir = "components"
const SrcDir = "src"

func main() {
	logger.InitShared(logger.LevelDebug)

	start := time.Now()

	// TODO: add a function to logger which has a cool processing animation or something
	logger.Info("Discovering documents...")
	scanned, err := discovery.DocumentDiscovery(SrcDir)
	if err != nil {
		logger.Error("Could not scan documents : %s", err.Error())
		return
	}

	componentPath := filepath.Join(SrcDir, ComponentsDir)
	logger.Info("Discovering components...")
	components, err := discovery.ComponentDiscovery(componentPath)
	if err != nil {
		logger.Error("Could not scan components : %s", err.Error())
		return
	}

	//fmt.Println(components)

	componentCache := caching.ComponentCache{
		Static:  make(map[string]*caching.Component),
		Dynamic: make(map[string]*caching.Component),
	}

	for _, filePath := range scanned.HtmlFiles {
		content, err := os.ReadFile(filePath)
		if err != nil {
			logger.Error("Could not read file %s : %s", filePath, err.Error())
			return
		}

		logger.Debug("File %s : %s", filePath, string(content))

		doc, err := html.Parse(bytes.NewReader(content))
		if err != nil {
			logger.Error("Could not parse file %s : %s", filePath, err.Error())
			return
		}

		var toReplace []*html.Node

		for node := range doc.Descendants() {
			if node.Type == html.ElementNode {
				tag := strings.ToLower(node.Data)

				if !htmlTags[tag] {
					_, dynamicExists := componentCache.Dynamic[tag]
					_, staticExists := componentCache.Static[tag]

					if !(dynamicExists || staticExists) && tag != "lua" {
						componentSrc, exists := components[tag]
						if !exists {
							logger.Info("Non-standard tag found in HTML and no component present : %s, assuming JS tag", tag)
							continue
						}

						logger.Info("Processing and caching tag %s...", tag)
						c, dynamic, err := caching.Cache(componentPath, componentSrc)
						if err != nil {
							logger.Error("Could not cache component %s : %s", componentSrc, err.Error())
							return
						}

						if dynamic {
							componentCache.Dynamic[tag] = c
						} else {
							componentCache.Static[tag] = c
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

		for _, node := range toReplace {
			stcComponent, staticExists := componentCache.Static[node.Data]
			dynComponent, dynamicExists := componentCache.Dynamic[node.Data]

			if staticExists {
				parent := node.Parent
				if parent != nil {
					for child := stcComponent.Node; child != nil; child = child.NextSibling {
						parent.InsertBefore(htmlUtilities.Clone(child), node)
					}
					parent.RemoveChild(node)
				}
			} else if dynamicExists {
				fmt.Println(dynComponent)
				logger.Warning("Dynamic components are not implemented yet, skipping %s...", node.Data)
				continue
			} else if node.Data == "lua" {
				logger.Warning("Lua components for regular input files are not implemented yet, skipping...")
				continue
			} else {
				logger.Info("Component %s not in cache, assuming JS tag and skipping...", node.Data)
				continue
			}
		}

		newWriter := bytes.NewBuffer(nil)
		err = html.Render(newWriter, doc)
		if err != nil {
			logger.Error("Could not render output : %s", err.Error())
			return
		}

		relPath, err := filepath.Rel(SrcDir, filePath)
		if err != nil {
			logger.Error("Could not get relative path : %s", err.Error())
			return
		}

		outPath := filepath.Join("build", relPath)
		_ = os.MkdirAll(filepath.Dir(outPath), 0755)

		err = os.WriteFile(outPath, newWriter.Bytes(), 0644)
		if err != nil {
			logger.Error("Could not write output : %s", err.Error())
			return
		}

		logger.Info("Saved to %s", outPath)
		logger.Debug("Output : %s", string(newWriter.Bytes()))
	}

	logger.Info("Finished in %s", time.Since(start))
}
