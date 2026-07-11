package config

import (
	"encoding/json"
	"os"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DB_url            string `json:"db_url"`
	Current_user_name string `json:"current_user_name"`
}

func Read() (Config, error) {
	config := Config{}
	filePath, err := getConfigFilePath()
	if err != nil {
		return config, nil
	}
	configFile, err := os.ReadFile(filePath)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func (cfg Config) SetUser(user string) error {
	cfg.Current_user_name = user
	err := write(cfg)
	if err != nil {
		return err
	}
	return nil
}

func getConfigFilePath() (string, error) {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return homePath + "/" + configFileName, nil
}

func write(cfg Config) error {
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	filePath, err := getConfigFilePath()
	if err != nil {
		return err
	}
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return err
	}
	return nil
}
