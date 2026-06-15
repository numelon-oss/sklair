package building

import (
	"fmt"
	"os"
	"path/filepath"
	"sklair/util"

	"github.com/numelon-oss/go-logger"
)

func copyStatic(files []plannedFile) error {
	logger.Info("Copying static files...")

	for _, file := range files {
		if err := os.MkdirAll(filepath.Dir(file.output), 0755); err != nil {
			return fmt.Errorf("could not create output directory for %s : %s", file.source, err.Error())
		}
		if err := util.CopyFile(file.source, file.output, 0644); err != nil {
			return fmt.Errorf("could not copy static file %s : %s", file.source, err.Error())
		}

		logger.Info("Copied static file to %s", file.output)
	}

	return nil
}
