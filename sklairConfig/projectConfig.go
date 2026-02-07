package sklairConfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sklair/constants"

	"github.com/invopop/jsonschema"
)

// TODO: expand this struct when JS obfuscation is added
type ObfuscateJS struct {
	// Whether JavaScript files should be obfuscated.
	//
	// TODO: this is a future feature
	Enabled bool `json:"enabled,omitempty" jsonschema:"title=Obfuscate JS"`
}

type PreventFOUC struct {
	// Whether sklair should help prevent Flash Of Unstyled Content (FOUC).
	Enabled bool `json:"enabled,omitempty" jsonschema:"title=Prevent FOUC"`

	// The colour of the FOUC prevention overlay, shown before the page is fully loaded.
	Colour string `json:"colour,omitempty" jsonschema:"title=FOUC prevention overlay colour"`
}

type HTTPMethod string

func (HTTPMethod) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type: "string",
		Enum: []any{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"HEAD",
			"OPTIONS",
		},
	}
}

type HooksHttpOptions struct {
	// Whether HTTP requests should be allowed in pre- / post-build hooks.
	HttpAllowed bool `json:"httpAllowed,omitempty" jsonschema:"title=Allow HTTP requests"`
	// A list of hosts that are allowed to make HTTP(s) requests in pre/post-build hooks.
	AllowedHosts []string `json:"allowedHosts,omitempty" jsonschema:"title=Allowed hosts"`
	// A list of methods that are allowed to be used in pre- / post-build hooks.
	AllowedMethods []HTTPMethod `json:"allowedMethods,omitempty" jsonschema:"title=Allowed methods"`

	// The maximum size of a response that can be received from a hook, in bytes.
	MaxResponseBytes int64 `json:"maxResponseBytes,omitempty" jsonschema:"title=Maximum response size"`
	// The maximum time in milliseconds that a HTTP(s) request within a hook can take to respond.
	Timeout int `json:"timeout,omitempty" jsonschema:"title=Timeout"`
	// Whether HTTP(s) redirects should be followed within a hook.
	FollowRedirects bool `json:"followRedirects,omitempty" jsonschema:"title=Follow redirects"`
	// The maximum number of redirects that can be followed by a HTTP(s) request within a hook.
	MaxRedirects int `json:"maxRedirects,omitempty" jsonschema:"title=Maximum redirects"`
}

type Hooks struct {
	// Whether sandboxed Lua pre- / post-build hooks should be executed.
	Enabled bool `json:"enabled,omitempty" jsonschema:"title=Enable hooks"`
	// The directory where pre- / post-build hooks are stored. The directory must contain the directories "pre" and "post".
	Path string `json:"path,omitempty" jsonschema:"title=Hooks directory"`

	// HTTP(s) request options for pre- / post-build hooks.
	Http *HooksHttpOptions `json:"http,omitempty" jsonschema:"title=HTTP options"`
}

//type ResourceHints struct {
//	Enabled    bool   `json:"enabled,omitempty"`
//	SiteOrigin string `json:"siteOrigin,omitempty"`
//}

// ProjectConfig (sklair.json) is Sklair's configuration file for each project.
type ProjectConfig struct {
	// Sklair configuration file schema.
	// This is usually automatically populated with whichever Sklair version created the project.
	SchemaURL string `json:"$schema,omitempty"`

	// Sandboxed Lua pre/post-build hook options.
	Hooks *Hooks `json:"hooks,omitempty" jsonschema:"title=Hooks"`

	// The directory where the project's source code is stored.
	Input string `json:"input,omitempty" jsonschema:"title=Input directory"`
	// The directory where the project's components are stored.
	Components string `json:"components,omitempty" jsonschema:"title=Components directory"`

	// A list of gitignore-style glob patterns that should be excluded from the build process.
	Exclude []string `json:"exclude,omitempty" jsonschema:"title=Exclude patterns"`
	// A list of gitignore-style glob patterns that should be excluded from the compilation process.
	ExcludeCompile []string `json:"excludeCompile,omitempty" jsonschema:"title=Exclude patterns for compiling"`

	// The directory where the built project should be written to.
	Output string `json:"output,omitempty" jsonschema:"title=Output directory"`

	// Whether HTML files should be minified during the build process.
	//
	// This field does not affect output at the moment as it is intended for the future.
	Minify bool `json:"minify,omitempty" jsonschema:"title=Minify HTML"`
	// Options for JavaScript obfuscation during the build process.
	//
	// This field does not affect output at the moment as it is intended for the future.
	ObfuscateJS *ObfuscateJS `json:"obfuscateJS,omitempty" jsonschema:"title=Obfuscate JavaScript"`

	// Options for preventing Flash Of Unstyled Content (FOUC) in the final outputted HTML.
	PreventFOUC *PreventFOUC `json:"preventFOUC,omitempty" jsonschema:"title=Prevent FOUC"`
	//ResourceHints *ResourceHints `json:"resourceHints,omitempty"` // TODO: in sklair init, add ResourceHints to the questionnaire
}

var DefaultConfig = ProjectConfig{
	Hooks: &Hooks{
		Enabled: false,
		Path:    "hooks",

		Http: &HooksHttpOptions{
			HttpAllowed:      false,
			AllowedHosts:     nil,
			AllowedMethods:   nil,
			MaxResponseBytes: 2 * 1024 * 1024, // 2 MiB
			Timeout:          5000,            // milliseconds
			FollowRedirects:  true,
			MaxRedirects:     5,
		},
	},

	Input:      "./src",
	Components: "./components",

	Exclude:        []string{},
	ExcludeCompile: []string{},

	Output: "./build",

	Minify: false,
	ObfuscateJS: &ObfuscateJS{
		Enabled: false,
	},

	PreventFOUC: &PreventFOUC{
		Enabled: false,
		Colour:  "#202020",
	},
	//ResourceHints: &ResourceHints{
	//	Enabled:    false,
	//	SiteOrigin: "https://sklair.numelon.com", // TODO: maybe just make it empty by default
	//},
}

func resolveProjectConfigPath() string {
	if _, err := os.Stat("sklair.json"); err == nil {
		return "sklair.json"
	}

	if _, err := os.Stat("src/sklair.json"); err == nil {
		return "src/sklair.json"
	}

	return "sklair.json" // default
}

func LoadProjectConfig() (*ProjectConfig, string, error) {
	configPath := resolveProjectConfigPath()

	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, "", err
	}

	config := DefaultConfig
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, "", err
	}

	return &config, filepath.Dir(configPath), nil
}

var SchemaURL string

func init() {
	regularVersion := constants.Version
	versionPath := regularVersion

	if regularVersion == "development" {
		versionPath = "latest/download"
		regularVersion = "development"
	} else {
		versionPath = "download/" + versionPath
	}

	SchemaURL = "https://github.com/numelon-oss/sklair/releases/" + versionPath + "/sklair-" + regularVersion + ".schema.json" // TODO: change to sklair.numelon.com

	DefaultConfig.SchemaURL = SchemaURL
}
