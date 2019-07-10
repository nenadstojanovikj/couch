package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/nenad/couch/pipeline"
	"github.com/nenad/couch/pkg/config"
	"github.com/nenad/couch/pkg/download"
	"github.com/nenad/couch/pkg/magnet"
	"github.com/nenad/couch/pkg/media"
	"github.com/nenad/couch/pkg/notifications"
	"github.com/nenad/couch/pkg/refresh"
	"github.com/nenad/couch/pkg/storage"
	"github.com/nenad/rd"
	"github.com/nenad/trakt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/streadway/handy/retry"
)

func NewAppCommand(config *config.Config, db *sql.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Run:   run(config, db),
		Short: "Runs the application",
		Long:  "Starts a daemon that will download files",
	}
}

func run(config *config.Config, db *sql.DB) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		repo := storage.NewMediaRepository(db)
		stop := make(chan os.Signal)
		signal.Notify(stop, os.Interrupt, os.Kill, syscall.SIGTERM)
		logrus.SetOutput(os.Stdout)
		logrus.SetLevel(logrus.DebugLevel)

		server := web.NewWebServer(config)
		notifier := newNotifier(config, db)

		searchItems := pipeline.NewPollStep(repo, pollers(config, repo)).
			Poll()
		magnetChan := pipeline.NewScrapeStep(repo, scrapers(config, repo)).
			Scrape(searchItems)
		downloadLocations := pipeline.NewExtractStep(repo, extractor(config, repo), config).
			Extract(magnetChan)
		downloadedItems := pipeline.NewDownloadStep(repo, downloader(config, repo), config.ConcurrentDownloadFiles, notifier).
			Download(downloadLocations)

		// Periodic pollers
		go func() {
			for {
				if err := refresh.Extract(repo, magnetChan); err != nil {
					logrus.Errorf("could not refresh database extract: %s", err)
				}

				time.Sleep(time.Minute * 10)
			}
		}()

		go func() {
			for {
				if err := refresh.Download(repo, downloadLocations); err != nil {
					logrus.Errorf("could not refresh database for downloads: %s", err)
				}

				time.Sleep(time.Minute * 10)
			}
		}()

		// Download notifier
		go func() {
			for item := range downloadedItems {
				logrus.Debugf("Downloaded %q", item.Term)
			}
		}()

		<-stop
	}
}

func newNotifier(c *config.Config, db *sql.DB) notifications.Notifier {
	if c.TelegramBotToken == "" {
		return &notifications.NoopNotifier{}
	}

	bot, err := tgbotapi.NewBotAPI(c.TelegramBotToken)
	if err != nil {
		logrus.Fatalf("error while creating Telegram Bot: %s", err)
	}

	client := notifications.NewTelegramClient(bot, db)
	go func() {
		if err := client.StartListener(); err != nil {
			logrus.Errorf("could not start Telegram listener: %s", err)
		}
	}()

	return client
}

func scrapers(c *config.Config, r *storage.MediaRepository) []magnet.Scraper {
	rarbgScraper, err := magnet.NewRarbgScraper()
	if err != nil {
		logrus.Fatalf("could not initialize rarbg: %s", err)
	}

	return []magnet.Scraper{
		rarbgScraper,
	}
}

func pollers(c *config.Config, r *storage.MediaRepository) []media.Provider {
	client := &http.Client{}
	client.Transport = retry.Transport{
		Next:  http.DefaultTransport,
		Delay: retry.Exponential(time.Second),
		Retry: retry.All(
			retry.Timeout(time.Second*10),
			retry.Errors(),
			retry.Temporary(),
			retry.Over(399),
		),
	}

	traktClient := trakt.NewClient(
		c.Trakt.ClientID,
		c.Trakt.ClientSecret,
		createTraktToken(&c.Trakt),
		client,
		nil,
	)

	return []media.Provider{
		media.NewTraktProvider(traktClient),
	}
}

func extractor(c *config.Config, r *storage.MediaRepository) magnet.Extractor {
	switch c.Downloader {
	case download.TypeHTTP:
		client := &http.Client{}
		client.Transport = retry.Transport{
			Next:  http.DefaultTransport,
			Delay: retry.Exponential(time.Second),
			Retry: retry.All(
				retry.Timeout(time.Second*10),
				retry.Errors(),
				retry.Temporary(),
				retry.Over(399),
			),
		}

		return magnet.NewRealDebridExtractor(
			rd.NewRealDebrid(createToken(&c.RealDebrid), client, rd.AutoRefresh),
			time.Second*10,
			false,
		)
	case download.TypeTorrent:
		return magnet.NewTorrentExtractor(r)
	default:
		panic(fmt.Errorf("extractor %s not found", c.Downloader))
	}
}

func downloader(c *config.Config, r *storage.MediaRepository) download.Getter {
	switch c.Downloader {
	case download.TypeTorrent:
		return download.NewTorrentDownloader(r, c)
	case download.TypeHTTP:
		return download.NewHttpDownloader()
	default:
		panic(fmt.Errorf("downloader %s not found", c.Downloader))
	}
}

func createToken(conf *config.AuthConfig) rd.Token {
	return rd.Token{
		AccessToken:  conf.AccessToken,
		TokenType:    conf.TokenType,
		ExpiresIn:    conf.ExpiresIn,
		ObtainedAt:   conf.ObtainedAt,
		RefreshToken: conf.RefreshToken,
	}
}

func createTraktToken(conf *config.AuthConfig) trakt.Token {
	return trakt.Token{
		AccessToken:  conf.AccessToken,
		TokenType:    conf.TokenType,
		ExpiresIn:    int64(conf.ExpiresIn),
		CreatedAt:    conf.ObtainedAt.Unix(),
		RefreshToken: conf.RefreshToken,
		Scope:        "public",
	}
}
