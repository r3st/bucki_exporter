package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

// Config struct for Ports and URLs from cmdline or config file
type Config struct {
	Port       string
	MetricsPath string
	ConfigPath string
	ClientTimeout int
	Urls       []string
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
