package building

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sklair/building/priorities"
	"sklair/caching"
	"sklair/devserver"
	"sklair/discovery"
	"sklair/htmlUtilities"
	"sklair/logger"
	"sklair/sklairConfig"
	"sklair/snippets"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func Build(config *sklairConfig.ProjectConfig, configDir string, outputDirOverride string) error {
	inputDir := filepath.Join(configDir, config.Input)
	componentsDir := filepath.Join(configDir, config.Components)

	outputDir := outputDirOverride
	if outputDirOverride == "" {
		outputDir = filepath.Join(configDir, config.Output)
	}

	start := time.Now()

	logger.Info("Indexing documents...")
	excludes := append(config.Exclude, config.Components, config.Output)
	scanned, err := discovery.DiscoverDocuments(inputDir, excludes)
	if err != nil {
		return errors.New("could not scan documents : " + err.Error())
	}

	logger.Info("Indexing components...")
	components, err := discovery.DiscoverComponents(componentsDir)
	if err != nil {
		return errors.New("could not scan components : " + err.Error())
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
			return errors.New("could not get PreventFOUC nodes : " + err.Error())
		}
	}

	logger.Info("Resolving components usage and compiling...")
	for _, filePath := range scanned.HtmlFiles {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("could not read file %s : %s", filePath, err.Error())
		}

		//logger.Debug("File %s : %s", filePath, string(content))

		doc, err := html.Parse(bytes.NewReader(content))
		if err != nil {
			return fmt.Errorf("could not parse file %s : %s", filePath, err.Error())
		}

		var toReplace []*html.Node

		for node := range doc.Descendants() {
			if node.Type == html.ElementNode {
				tag := strings.ToLower(node.Data)

				if !htmlUtilities.HtmlTags[tag] {
					_, dynamicExists := componentCache.Dynamic[tag]
					_, staticExists := componentCache.Static[tag]

					if !(dynamicExists || staticExists) && (!(tag == "lua" || tag == "opengraph")) {
						componentSrc, exists := components[tag]
						if !exists {
							logger.Warning("Non-standard tag found in HTML and no component present : %s, assuming JS tag", tag)
							continue
						}

						logger.Info("Processing and caching tag %s...", tag)
						cached, err := caching.MakeCache(componentsDir, componentSrc)
						if err != nil {
							return fmt.Errorf("could not cache component %s : %s", componentSrc, err.Error())
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
		body := htmlUtilities.FindTag(doc, "body")
		if head == nil || body == nil {
			return fmt.Errorf("could not find head or body tags in %s, how does that even happen", filePath)
		}

		// usedComponents ensures that each component contributes its <head> nodes at most ONCE per document,
		// even if the component appears multiple times in the source document
		usedComponents := make(map[string]struct{})
		// seenHead, on the other hand, is used for actual deduplication
		for _, originalTag := range toReplace {
			stcComponent, staticExists := componentCache.Static[originalTag.Data]
			dynComponent, dynamicExists := componentCache.Dynamic[originalTag.Data]

			parent := originalTag.Parent
			if parent == nil {
				return fmt.Errorf("somehow the parent does not exist for %s. (memory corruption???)", originalTag.Data)
			}

			//fmt.Println(originalTag.Data)

			// TODO: the logic for static and dynamic components will likely be very similar
			// in the future, simply combine both branches,
			// but for dynamic components just have a simple processing stage.
			// after that its treated as a static component would be
			if staticExists {
				htmlUtilities.InsertNodesBefore(originalTag, stcComponent.BodyNodes)

				// this check ensures that each component contributes its <head> nodes at most ONCE per document
				if _, seen := usedComponents[originalTag.Data]; !seen {
					htmlUtilities.AppendNodes(head, stcComponent.HeadNodes)
				}
				usedComponents[originalTag.Data] = struct{}{}
				parent.RemoveChild(originalTag)
			} else if dynamicExists {
				fmt.Println(dynComponent)
				logger.Warning("Dynamic components are not implemented yet, skipping %s...", originalTag.Data)
				continue
			} else if originalTag.Data == "lua" {
				// TODO: prints from lua will be appended to a buffer
				// then this buffer will be parsed by html
				// then this will be inserted into document
				logger.Warning("Lua components for regular input files are not implemented yet, skipping...")
				continue
			} else if originalTag.Data == "opengraph" {
				for _, child := range snippets.OpenGraph(originalTag) {
					head.AppendChild(child)
				}
				parent.RemoveChild(originalTag)
			} else {
				logger.Warning("Component %s not in cache, assuming JS tag and skipping...", originalTag.Data) // TODO: specify what a JS tag actually is
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

	processingEnd := time.Since(start)

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

		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("could not read static file %s : %s", filePath, err.Error())
		}

		err = os.WriteFile(outPath, data, 0644)
		if err != nil {
			return fmt.Errorf("could not write static file %s : %s", filePath, err.Error())
		}

		logger.Info("Copied static file to %s", outPath)
	}

	//logger.EmptyLine()
	logger.Info("Compilation time (including writes) : %s", processingEnd)
	logger.Info("Static copy time : %s", time.Since(staticStart))
	logger.Info("Total processing time : %s", time.Since(start))

	return nil
}
