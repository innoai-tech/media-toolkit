package livestream

import (
	"encoding/json"
	"os"
)

func LoadStreams(configFile string) ([]Stream, error) {
	jsonRaw, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	streams := make([]Stream, 0)
	if err = json.Unmarshal(jsonRaw, &streams); err != nil {
		return nil, err
	}
	return streams, nil
}
