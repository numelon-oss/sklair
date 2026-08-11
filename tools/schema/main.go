package main

import (
	"encoding/json"
	"os"
	"sklair/sklairConfig"
	"strings"

	"github.com/invopop/jsonschema"
)

func main() {
	r := &jsonschema.Reflector{
		AllowAdditionalProperties:  false,
		RequiredFromJSONSchemaTags: true,
	}

	if err := r.AddGoComments("sklair", "./sklairConfig"); err != nil {
		panic(err)
	}

	schema := r.Reflect(&sklairConfig.ProjectConfig{})
	applyDefaults(schema, "ProjectConfig", projectDefaults())

	schema.Title = "Sklair Project Configuration"
	schema.ID = jsonschema.ID(sklairConfig.SchemaURL)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")
	_ = enc.Encode(schema)
}

func projectDefaults() map[string]any {
	defaults := sklairConfig.DefaultConfig
	return map[string]any{
		"$schema":        defaults.SchemaURL,
		"input":          defaults.Input,
		"components":     defaults.Components,
		"layouts":        defaults.Layouts,
		"exclude":        []string{},
		"excludeCompile": []string{},
		"output":         defaults.Output,
		"minify":         defaults.Minify,
		"templates":      []string{},
		"hooks": map[string]any{
			"enabled": defaults.Hooks.Enabled,
			"path":    defaults.Hooks.Path,
			"http": map[string]any{
				"httpAllowed":      defaults.Hooks.Http.HttpAllowed,
				"allowedHosts":     []string{},
				"allowedMethods":   []sklairConfig.HTTPMethod{},
				"maxResponseBytes": defaults.Hooks.Http.MaxResponseBytes,
				"timeout":          defaults.Hooks.Http.Timeout,
				"followRedirects":  defaults.Hooks.Http.FollowRedirects,
				"maxRedirects":     defaults.Hooks.Http.MaxRedirects,
			},
		},
		"obfuscateJS": map[string]any{
			"enabled": defaults.ObfuscateJS.Enabled,
		},
		"preventFOUC": map[string]any{
			"enabled": defaults.PreventFOUC.Enabled,
			"colour":  defaults.PreventFOUC.Colour,
		},
	}
}

func applyDefaults(root *jsonschema.Schema, definitionName string, defaults map[string]any) {
	definition := root.Definitions[definitionName]
	if definition == nil || definition.Properties == nil {
		return
	}
	for name, value := range defaults {
		property, exists := definition.Properties.Get(name)
		if !exists {
			continue
		}
		property.Default = value
		object, objectDefault := value.(map[string]any)
		if !objectDefault || property.Ref == "" {
			continue
		}
		applyDefaults(root, strings.TrimPrefix(property.Ref, "#/$defs/"), object)
	}
}
