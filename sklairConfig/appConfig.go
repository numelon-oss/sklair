package sklairConfig

// TODO: check TODO.md for more info about this

type AppConfig struct {
	AutoUpdate bool `json:"autoUpdate,omitempty"`
}

var defaultAppConfig = AppConfig{
	AutoUpdate: true,
}

//func Load(path string) (*ProjectConfig, error) {
//	file, err := os.ReadFile(path)
//	if err != nil {
//		return nil, err
//	}
//
//	config := DefaultConfig
//	if err := json.Unmarshal(file, &config); err != nil {
//		return nil, err
//	}
//
//	return &config, nil
//}
