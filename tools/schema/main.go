package main

import (
	"encoding/json"
	"os"
	"sklair/sklairConfig"

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

	schema.Title = "Sklair Project Configuration"
	schema.ID = jsonschema.ID(sklairConfig.SchemaURL) // TODO: change to sklair.numelon.com

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")
	_ = enc.Encode(schema)
}
