package sklairConfig

import (
	"os"
	"path/filepath"
)

// TODO: check TODO.md for more info about this

type GlobalConfig struct {
	// TODO: to be removed, as Sklair is managed by package managers
	//CheckForUpdates bool `json:"checkForUpdates,omitempty"`
}

var defaultGlobalConfig = GlobalConfig{
	//CheckForUpdates: true,
}

func GlobalConfigPath() (string, error) {
	home, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, "sklair/config.json"), nil
}
