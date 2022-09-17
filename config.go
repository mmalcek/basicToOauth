package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type tConfig struct {
	Host         string   `yaml:"host"`
	Port         string   `yaml:"port"`
	ClientID     string   `yaml:"client_id"`
	TenantID     string   `yaml:"tenant_id"`
	ProxyURL     string   `yaml:"proxy_url"`
	AuthorityURL string   `yaml:"authority_url"`
	Scopes       []string `yaml:"scopes"`
}

var config tConfig

func loadConfig() error {
	file, err := os.ReadFile("config.yaml")
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return err
	}
	return nil
}
