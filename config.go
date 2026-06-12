package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	Host               string              `json:"host"`
	Port               int                 `json:"port"`
	Username           string              `json:"username"`
	Password           string              `json:"password"`
	Database           string              `json:"database"`
	Table              string              `json:"table"`
	CSVBaseName        string              `json:"csv_base_name"`
	IdsToFields        map[string]string   `json:"ids_to_fieds"`
	PostTypeToMetaKeys map[string][]string `json:"post_type_to_meta_keys"`
}

func LoadConfig(filename string) (*Config, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
