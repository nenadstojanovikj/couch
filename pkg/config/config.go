package config

import (
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
	Downloader string `json:"downloader"`

	Port int `json:"port"`

	MoviesPath  string `json:"movies_path"`
	TVShowsPath string `json:"tvshows_path"`

	ConcurrentDownloadFiles int `json:"concurrent_download_files"`

	RealDebrid AuthConfig `json:"real_debrid"`
	Trakt      AuthConfig `json:"trakt_tv"`

	TelegramBotToken string `json:"telegram_bot_token"`
}

func InitConfiguration(store Loader) (conf Config, err error) {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}

	val, err := store.Load()
	if err != nil {
		return Config{
			Downloader:              "http",
			Port:                    8080,
			MoviesPath:              u.HomeDir + "/Movies",
			TVShowsPath:             u.HomeDir + "/TVShows",
			ConcurrentDownloadFiles: 3,
		}, err
	}

	if c, ok := val.(Config); !ok {
		return c, fmt.Errorf("invalid config store: %#v", val)
	} else {
		conf = c
	}

	return conf, nil
}
