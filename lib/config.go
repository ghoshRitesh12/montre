package lib

import (
	"encoding/json"
	"errors"
	"log"
	"os"
)

type Config struct {
	MainFile   string   `json:"mainFile"`
	WatchExts  []string `json:"watchExtensions"`
	IgnoreDirs []string `json:"ignoreDirs"`
}

const ConfigFileName string = "montre.json"

// to check if config file exists
func configFileExist() bool {
	_, err := os.Stat(ConfigFileName)
	return !errors.Is(err, os.ErrNotExist)
}

func populateConfig(m *Montre) {
	config := Config{}

	configFileBytes, err := os.ReadFile(ConfigFileName)
	if err != nil {
		log.Fatalln(ErrReadingConfigFile)
	}

	// unmarshaling data from config json file
	if err := json.Unmarshal(configFileBytes, &config); err != nil {
		log.Fatalln(ErrParsingConfigFile)
	}

	if m.config.MainFile == "" && config.MainFile == "" {
		log.Fatalln(ErrNoMainFile)
	}

	if len(config.WatchExts) < 1 {
		config.WatchExts = []string{".go"}
	}

	config.IgnoreDirs = append([]string{".git"}, config.IgnoreDirs...)
	m.config = config
}
