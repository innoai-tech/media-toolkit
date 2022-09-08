package core

import (
	"encoding/json"
	"os"
)

type Stream struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Rtsp string `json:"rtsp"`
}

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
