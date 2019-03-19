package config

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"os/user"
	"time"
)

type AuthConfig struct {
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int64     `json:"expires_in"`
	ObtainedAt   time.Time `json:"obtained_at"`
	TokenType    string    `json:"token_type"`
}

type Config struct {
	db *sql.DB

	Downloader string `json:"downloader"`

	Port int `json:"port"`

	MoviesPath  string `json:"movies_path"`
	TVShowsPath string `json:"tvshows_path"`

	ConcurrentDownloadFiles int `json:"concurrent_download_files"`

	RealDebrid AuthConfig `json:"real_debrid"`
	Trakt      AuthConfig `json:"trakt_tv"`

	TelegramBotToken string `json:"telegram_bot_token"`
}

func NewConfiguration(db *sql.DB) Config {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}

	return Config{
		db: db,

		Downloader:              "http",
		Port:                    8080,
		MoviesPath:              u.HomeDir + "/Movies",
		TVShowsPath:             u.HomeDir + "/TVShows",
		ConcurrentDownloadFiles: 3,
	}
}

func (c *Config) Load() error {
	row := c.db.QueryRow("SELECT config FROM config LIMIT 1;")

	var j []byte
	err := row.Scan(&j)

	if err != nil {
		return err
	}

	if bytes.Equal(j, []byte("{}")) {
		return fmt.Errorf("empty config")
	}

	return json.Unmarshal(j, &c)
}

func (c *Config) Save() error {
	// Indenting so it's human readable for easier inspection
	b, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}

	_, err = c.db.Exec("UPDATE config SET config = ?", b)
	return err
}
