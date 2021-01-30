package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

// Checks saves names of checks and names their data, needed for deletion of obsolete checkes or data
type checkData struct {
	Name string
    Data []string
}

// url describe given Url by addess and name from config
type url struct {
  Name string
  Address string
}

// Config struct for Ports and URLs from cmdline or config file
type Config struct {
	Port       string
	MetricsPath string
	ConfigPath string
	ClientTimeout int
	Urls       []url
	BuckiMetrics bool
}

// ReadConfig get configuration from config yaml
func ReadConfig(cfg *Config) {
	fileContent, err := ioutil.ReadFile(cfg.ConfigPath)

	if err != nil {
		log.Fatal(err)
	}

	err = yaml.Unmarshal(fileContent, &cfg)

	if err != nil {
		log.Fatal(err)
	}
}
