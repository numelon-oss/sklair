package schema

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

const Draft2020 = "https://json-schema.org/draft/2020-12/schema"

type Validator struct {
	compiled *jsonschema.Schema
}

type Issue struct {
	InstancePath string
	SchemaPath   string
	Message      string
}

func Compile(document any, location string, loader jsonschema.URLLoader) (*Validator, error) {
	if err := CheckDocument(document); err != nil {
		return nil, err
	}
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	compiler.AssertFormat()
	compiler.UseLoader(loader)
	if err := compiler.AddResource(location, document); err != nil {
		return nil, err
	}
	compiled, err := compiler.Compile(location)
	if err != nil {
		return nil, err
	}
	return &Validator{compiled: compiled}, nil
}

func CheckDocument(document any) error {
	object, ok := document.(map[string]any)
	if !ok {
		return nil
	}
	declared, exists := object["$schema"]
	if !exists {
		return nil
	}
	name, ok := declared.(string)
	if !ok {
		return fmt.Errorf("JSON Schema $schema must be a string")
	}
	if strings.TrimSuffix(name, "#") != Draft2020 {
		return fmt.Errorf("unsupported JSON Schema draft %q; Sklair supports Draft 2020-12", name)
	}
	return nil
}

func (v *Validator) Validate(value any) ([]Issue, error) {
	err := v.compiled.Validate(value)
	if err == nil {
		return nil, nil
	}
	var validationError *jsonschema.ValidationError
	if !errors.As(err, &validationError) {
		return nil, err
	}
	output := validationError.BasicOutput()
	issues := make([]Issue, 0)
	collectIssues(*output, &issues)
	sort.Slice(issues, func(i, j int) bool {
		if issues[i].InstancePath != issues[j].InstancePath {
			return issues[i].InstancePath < issues[j].InstancePath
		}
		if issues[i].SchemaPath != issues[j].SchemaPath {
			return issues[i].SchemaPath < issues[j].SchemaPath
		}
		return issues[i].Message < issues[j].Message
	})
	return issues, nil
}

func collectIssues(output jsonschema.OutputUnit, issues *[]Issue) {
	if output.Error != nil {
		*issues = append(*issues, Issue{
			InstancePath: output.InstanceLocation,
			SchemaPath:   output.KeywordLocation,
			Message:      output.Error.String(),
		})
	}
	for _, child := range output.Errors {
		collectIssues(child, issues)
	}
}
