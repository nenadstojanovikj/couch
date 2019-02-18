package cmd

import (
	"context"
	"fmt"
	"github.com/nenadstojanovikj/couch/pipeline"
	"github.com/nenadstojanovikj/couch/pkg/config"
	"github.com/nenadstojanovikj/couch/pkg/download"
	"github.com/nenadstojanovikj/couch/pkg/magnet"
	"github.com/nenadstojanovikj/couch/pkg/mediaprovider"
	"github.com/nenadstojanovikj/couch/pkg/refresh"
	"github.com/nenadstojanovikj/couch/pkg/storage"
	"github.com/nenadstojanovikj/couch/pkg/web"
	"github.com/nenadstojanovikj/rd"
	"github.com/nenadstojanovikj/trakt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/streadway/handy/retry"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func NewAppCommand(config *config.Config, repo *storage.MediaRepository) *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Run:   run(config, repo),
		Short: "Runs the application",
		Long:  "Starts a web server and a daemon that will download files",
	}
}

func run(config *config.Config, repo *storage.MediaRepository) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		stop := make(chan os.Signal)
		signal.Notify(stop, os.Interrupt, os.Kill, syscall.SIGTERM)
		logrus.SetOutput(os.Stdout)
		logrus.SetLevel(logrus.DebugLevel)

		server := web.NewWebServer(config)
		go func() {
			if err := server.ListenAndServe(); err != nil {
				logrus.Errorf("web server failed to run: %s", err)
			}
		}()

		searchItems := pipeline.NewPollStep(repo, pollers(config, repo)).
			Poll()
		magnetChan := pipeline.NewScrapeStep(repo, scrapers(config, repo)).
			Scrape(searchItems)
		downloadLocations := pipeline.NewExtractStep(repo, extractor(config, repo), config).
			Extract(magnetChan)
		downloadedItems := pipeline.NewDownloadStep(repo, downloader(config, repo), config.ConcurrentDownloadFiles).
			Download(downloadLocations)

		// Periodic pollers
		go func() {
			for {
				if err := refresh.Extract(repo, magnetChan); err != nil {
					logrus.Error("could not refresh database extract: %s", err)
				}

				time.Sleep(time.Minute * 10)
			}
		}()

		go func() {
			for {
				if err := refresh.Download(repo, downloadLocations); err != nil {
					logrus.Error("could not refresh database for downloads: %s", err)
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
		_ = server.Shutdown(context.TODO())
	}
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

func pollers(c *config.Config, r *storage.MediaRepository) []mediaprovider.Poller {
	client := &http.Client{}
	client.Transport = retry.Transport{
		Next:  http.DefaultTransport,
		Delay: retry.Exponential(time.Second),
		Retry: retry.All(retry.Timeout(time.Second*30), retry.Errors(), retry.Over(399), retry.Method("GET", "POST")),
	}

	traktClient := trakt.NewClient(
		c.Trakt.ClientID,
		c.Trakt.ClientSecret,
		createTraktToken(&c.Trakt),
		client,
		nil,
	)

	return []mediaprovider.Poller{
		mediaprovider.NewTraktProvider(traktClient),
	}
}

func extractor(c *config.Config, r *storage.MediaRepository) magnet.Extractor {
	switch c.Downloader {
	case download.TypeHTTP:
		client := &http.Client{}
		client.Transport = retry.Transport{
			Next:  http.DefaultTransport,
			Delay: retry.Exponential(time.Second),
			Retry: retry.All(retry.Timeout(time.Second*30), retry.Errors(), retry.Temporary(), retry.Over(399), retry.Method("GET", "POST")),
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
