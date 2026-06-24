package providers

import (
	"encoding/json"
	"os"
)

func SaveRegistry(path string, registry Registry) error {
	bytes, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return err
	}
	bytes = append(bytes, '\n')
	return os.WriteFile(path, bytes, 0o640)
}
