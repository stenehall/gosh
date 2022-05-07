package config

import (
	"fmt"
	"log"
	"os"

	"github.com/stenehall/gosh/internal/favii"
	"gopkg.in/yaml.v3"
)

// Config is the main config struct.
type Config struct {
	filePath string
	Data     Gosh
}

// Gosh is the config files data structure.
type Gosh struct {
	Title      string
	ShowTitle  bool
	Background string
	Port       int
	Sets       []Set
}

// Set is a single section of the dashboard, each set can contain several sites.
type Set struct {
	Name  string
	Icon  string
	Sites []favii.Site
}

// New creates a new Favii struct with http.DefaultClient and empty map, also
// an optional cache map.
func New(filePath string) (*Config, error) {
	data, err := loadConfig(filePath)

	if err != nil {
		return nil, err
	}

	return &Config{
		filePath,
		data,
	}, nil
}

func loadConfig(configFile string) (Gosh, error) {
	gosh := Gosh{}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return Gosh{}, fmt.Errorf("failed to read the config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &gosh); err != nil {
		return Gosh{}, fmt.Errorf("error from unmarshalled yaml: %w", err)
	}

	return gosh, nil
}

// SaveConfig persists the current config to disk.
func (config *Config) SaveConfig() error {
	out, err := yaml.Marshal(config.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal configs file: %w", err)
	}

	if err := os.WriteFile(config.filePath, out, 0600); err != nil {
		return fmt.Errorf("failed to save updated config %w", err)
	}

	log.Println("saving updated config with favicons")
	return nil
}

// CountSets returns the number of sets in the current config.
func (config *Config) CountSets() int {
	return len(config.Data.Sets)
}

// CountSites returns the total number of sites in all sets in the current config.
func (config *Config) CountSites() int {
	sites := 0

	for _, set := range config.Data.Sets {
		sites += len(set.Sites)
	}
	return sites
}
