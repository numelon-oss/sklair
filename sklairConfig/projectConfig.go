package sklairConfig

import (
	"encoding/json"
	"os"
)

type PreventFOUC struct {
	Enabled bool   `json:"enabled,omitempty"`
	Colour  string `json:"colour,omitempty"`
}

type ProjectConfig struct {
	Input      string   `json:"input,omitempty"`
	Components string   `json:"components,omitempty"`
	Exclude    []string `json:"exclude,omitempty"`

	ExcludeCompile []string `json:"excludeCompile,omitempty"`
	Output         string   `json:"output,omitempty"`

	PreventFOUC *PreventFOUC `json:"PreventFOUC,omitempty"`

	Minify    bool `json:"minify,omitempty"`
	Obfuscate bool `json:"obfuscate,omitempty"`
}

var DefaultConfig = ProjectConfig{
	Input:      "./",
	Components: "./components",
	Exclude:    []string{},

	ExcludeCompile: []string{},
	Output:         "./build",

	PreventFOUC: &PreventFOUC{
		Enabled: false,
		Colour:  "#202020",
	},

	Minify:    false,
	Obfuscate: false, // this pertains to JS obfuscation, not HTML.. you cant really obfuscate HTML per se
	// TODO: likely rename "Obfuscate" to "ObfuscateJS" and make it another struct with more properties and customisation
}

func Load(path string) (*ProjectConfig, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := DefaultConfig
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
