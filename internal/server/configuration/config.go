package configuration

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	Certificates []Certificate `json:"certificates"`
}

type Certificate struct {
	CertPath    string `json:"cert_path"`
	CertKeyPath string `json:"cert_key_path"`
}

func LoadConfiguration(configPath string) (*Configuration, error) {
	fileBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config *Configuration
	if err = json.Unmarshal(fileBytes, &config); err != nil {
		return nil, err
	}

	return config, nil
}
