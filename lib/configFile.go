package lib

import (
	"encoding/json"
	"errors"
	"os"
)

type config struct {
	Main              string `json:"main"`
	AllowList         string `json:"allowList"`
	IgnoreList        string `json:"ignoreList"`
	IgnoreHiddenFiles bool   `json:"ignoreHiddenFiles"`
}

const ConfigFileName string = "montre.json"

// to check if config file exists
func configFileExist() bool {
	_, err := os.Stat(ConfigFileName)
	return !errors.Is(err, os.ErrNotExist)
}

func populateConfig(m *Montre) {
	config := config{}

	configFileBytes, err := os.ReadFile(ConfigFileName)
	if err != nil {
		panic(ErrReadingConfigFile)
	}

	if err := json.Unmarshal(configFileBytes, &config); err != nil {
		panic(ErrParsingConfigFile)
	}

	if m.Filename == "" || config.Main == "" {
		panic(ErrNoMainFile)
	}

	if config.AllowList == "" {
		panic(ErrEmptyAllowList)
	}

	m.Filename = config.Main
	m.AllowList = config.AllowList
	m.IgnoreList = config.IgnoreList
	m.IgnoreHiddenFiles = config.IgnoreHiddenFiles
}
