package sklairConfig

import (
	"encoding/json"
	"os"
)

type preventFOUC struct {
	Enabled bool   `json:"enabled,omitempty"`
	Colour  string `json:"colour,omitempty"`
}

type SklairConfig struct {
	PreventFOUC preventFOUC `json:"preventFOUC,omitempty"`

	Input      string   `json:"input,omitempty"`
	Components string   `json:"components,omitempty"`
	Exclude    []string `json:"exclude,omitempty"`

	ExcludeCompile []string `json:"excludeCompile,omitempty"`
	Output         string   `json:"output,omitempty"`

	Minify    bool `json:"minify,omitempty"`
	Obfuscate bool `json:"obfuscate,omitempty"`
}

var defaultConfig = SklairConfig{
	PreventFOUC: preventFOUC{
		Enabled: false,
	},
	Input:  "./",
	Output: "./build",

	Minify:    false,
	Obfuscate: false, // this pertains to JS obfuscation, not HTML.. you cant really ofuscate HTML per se
	// TODO: likely rename "Obfuscate" to "ObfuscateJS" and make it another struct with more properties and customisation
}

func Load(path string) (*SklairConfig, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := defaultConfig
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
