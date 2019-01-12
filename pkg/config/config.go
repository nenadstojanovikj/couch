package config

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type Configuration interface {
}

type AuthConfig struct {
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	ObtainedAt   time.Time `json:"obtained_at"`
	TokenType    string    `json:"token_type"`
}

type ShowRssConfig struct {
	PersonalFeed string `json:"personal_feed"`
}

type Config struct {
	configPath string
	mu         sync.Mutex

	MoviesPath  string `json:"movies_path"`
	TVShowsPath string `json:"tvshows_path"`

	MoviesInSeparateDirectories  bool          `json:"movies_in_separate_directories"`
	TVShowsInSeparateDirectories bool          `json:"tvshows_in_separate_directories"`
	MaximumDownloadSpeed         string        `json:"maximum_download_speed"`
	RealDebrid                   AuthConfig    `json:"real_debrid"`
	TraktTV                      AuthConfig    `json:"trakt_tv"`
	ShowRss                      ShowRssConfig `json:"showrss"`
	Port                         int           `json:"port"`
}

func NewConfiguration() Config {
	return Config{}
}

func (c *Config) LoadFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	c.configPath = path

	return json.NewDecoder(file).Decode(c)
}

func (c *Config) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	file, err := os.Create(c.configPath + ".tmp")
	if err != nil {
		return err
	}

	// Indenting so it's human readable for easier inspection
	bytes, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}
	if _, err = file.Write(bytes); err != nil {
		return err
	}

	return os.Rename(c.configPath+".tmp", c.configPath)
}
