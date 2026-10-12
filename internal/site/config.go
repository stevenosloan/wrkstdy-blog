package site

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// Config is site.toml.
type Config struct {
	BaseURL     string     `toml:"baseURL"`
	Title       string     `toml:"title"`
	Description string     `toml:"description"`
	Author      string     `toml:"author"`
	Language    string     `toml:"language"`
	Menu        []MenuItem `toml:"menu"`
}

type MenuItem struct {
	Name string `toml:"name"`
	URL  string `toml:"url"`
}

// Project is one entry in data/projects.toml, listed on the homepage.
type Project struct {
	Name        string `toml:"name"`
	URL         string `toml:"url"`
	Description string `toml:"description"`
}

func loadConfig(path string) (Config, error) {
	cfg := Config{Language: "en"}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg, fmt.Errorf("reading config: %w", err)
	}
	return cfg, nil
}

func loadProjects(path string) ([]Project, error) {
	var data struct {
		Projects []Project `toml:"projects"`
	}
	if _, err := toml.DecodeFile(path, &data); err != nil {
		return nil, fmt.Errorf("reading projects: %w", err)
	}
	return data.Projects, nil
}
