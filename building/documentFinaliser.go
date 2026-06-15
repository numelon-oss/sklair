package building

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sklair/building/priorities"
	"sklair/devserver"
	"sklair/discovery"
	"sklair/htmlUtilities"
	sklairRuntime "sklair/runtime"
	"sklair/sklairConfig"
	"sklair/snippets"
	"sort"

	"github.com/numelon-oss/go-logger"
	"golang.org/x/net/html"
)

type renderedDocument struct {
	plannedFile
	contents []byte
}

type buildOutput struct {
	documents        []renderedDocument
	componentFolders map[string]discovery.ComponentSource
	runtime          bool
	devServer        bool
}

func finaliseDocuments(documents []*documentState, config *sklairConfig.ProjectConfig, devServer bool) (*buildOutput, error) {
	var preventFoucHead *html.Node
	var preventFoucBody *html.Node
	if config.PreventFOUC != nil && config.PreventFOUC.Enabled {
		var err error
		preventFoucHead, preventFoucBody, err = snippets.GetFOUCNodes(config.PreventFOUC.Colour)
		if err != nil {
			return nil, fmt.Errorf("could not get PreventFOUC nodes : %s", err.Error())
		}
	}

	output := &buildOutput{
		documents:        make([]renderedDocument, 0, len(documents)),
		componentFolders: make(map[string]discovery.ComponentSource),
		devServer:        devServer,
	}

	for _, document := range documents {
		filePath := document.source
		doc := document.root
		head := htmlUtilities.FindTag(doc, "head")
		body := htmlUtilities.FindTag(doc, "body")
		if head == nil || body == nil {
			return nil, fmt.Errorf("could not find head or body tags in %s after compilation", filePath)
		}

		for name, source := range document.componentFolders {
			output.componentFolders[name] = source
		}

		if len(document.templates) > 0 {
			templateNames := make([]string, 0, len(document.templates))
			for name := range document.templates {
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
				component := document.templates[name]

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
			output.runtime = true
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

		segmentedHead, err := SegmentHead(head)
		if err != nil {
			return nil, fmt.Errorf("could not segment <head> in %s : %s", filePath, err.Error())
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

		if devServer {
			segmentedHead = append(segmentedHead, &HeadSegment{
				Nodes: []*html.Node{
					htmlUtilities.Clone(devserver.WSScriptNode),
				},
				TreatAsTag:        priorities.Script,
				IsOrderingBarrier: false,
			})
		}

		segmentedHead = OptimiseHead(segmentedHead)

		htmlUtilities.RemoveAllChildren(head)
		for _, segment := range segmentedHead {
			for _, node := range segment.Nodes {
				head.AppendChild(node)
			}
		}

		writer := bytes.NewBuffer(nil)
		if err := html.Render(writer, doc); err != nil {
			return nil, fmt.Errorf("could not render output for %s : %s", filePath, err.Error())
		}

		output.documents = append(output.documents, renderedDocument{
			plannedFile: document.plannedFile,
			contents:    writer.Bytes(),
		})
	}

	return output, nil
}

func replaceOutput(output *buildOutput, paths buildPaths) error {
	if err := os.RemoveAll(paths.output); err != nil {
		return fmt.Errorf("could not remove output directory %s : %s", paths.output, err.Error())
	}

	for _, document := range output.documents {
		if err := os.MkdirAll(filepath.Dir(document.output), 0755); err != nil {
			return fmt.Errorf("could not create output directory for %s : %s", document.source, err.Error())
		}
		if err := os.WriteFile(document.output, document.contents, 0644); err != nil {
			return fmt.Errorf("could not write output for %s : %s", document.source, err.Error())
		}

		logger.Info("Saved to %s", document.output)
	}

	if len(output.componentFolders) > 0 {
		logger.Info("Copying component assets...")
		if err := copyComponentFolders(paths.components, paths.output, output.componentFolders); err != nil {
			return err
		}
	}

	if output.runtime {
		path := filepath.Join(paths.output, filepath.FromSlash(sklairRuntime.OutputPath))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("could not create Sklair runtime directory : %s", err.Error())
		}
		if err := os.WriteFile(path, []byte(sklairRuntime.Module), 0644); err != nil {
			return fmt.Errorf("could not write Sklair runtime module : %s", err.Error())
		}
	}

	if output.devServer {
		if err := os.MkdirAll(filepath.Join(paths.output, "_sklair"), 0755); err != nil {
			return fmt.Errorf("could not create sklair dev server directory : %s", err.Error())
		}
		if err := os.WriteFile(filepath.Join(paths.output, devserver.WSDevScriptPath), []byte(devserver.WSDevScript), 0644); err != nil {
			return fmt.Errorf("could not write sklair dev server websocket js file : %s", err.Error())
		}
	}

	return nil
}
