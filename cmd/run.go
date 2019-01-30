package cmd

import (
	"bufio"
	"database/sql"
	"fmt"
	"github.com/nenadstojanovikj/couch/pipeline"
	"github.com/nenadstojanovikj/couch/pkg/config"
	"github.com/nenadstojanovikj/couch/pkg/download"
	"github.com/nenadstojanovikj/couch/pkg/magnet"
	"github.com/nenadstojanovikj/couch/pkg/media"
	"github.com/nenadstojanovikj/couch/pkg/refresh"
	"github.com/nenadstojanovikj/couch/pkg/storage"
	"github.com/nenadstojanovikj/rd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func NewAppCommand(config *config.Config, repo *storage.MediaRepository) *cobra.Command {
	return &cobra.Command{
		Use: "run",
		Run: run(config, repo),
	}
}

// Poll search providers -> search_items
// search_items -> Download torrents -> magnet_info
// magnet_info -> Extract links & files -> torrent_files & realdebrid
// torrent_files & realdebrid -> Download

// Utilize pipelines: https://blog.golang.org/pipelines
// https://medium.com/statuscode/pipeline-patterns-in-go-a37bb3a7e61d

func run(config *config.Config, repo *storage.MediaRepository) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		stop := make(chan os.Signal)
		signal.Notify(stop, os.Interrupt, os.Kill, syscall.SIGTERM)

		logrus.SetLevel(logrus.DebugLevel)

		searchItems := make(chan media.Metadata)

		go func() {
			for {
				reader := bufio.NewReader(os.Stdin)
				fmt.Println("Enter TV Show name:")
				name, _ := reader.ReadString('\n')
				fmt.Println("Season:")
				season, _ := reader.ReadString('\n')
				fmt.Println("Episode:")
				episode, _ := reader.ReadString('\n')
				name = strings.Trim(name, "\n")
				s, _ := strconv.Atoi(strings.Trim(season, "\n"))
				e, _ := strconv.Atoi(strings.Trim(episode, "\n"))
				if name != "" && s != 0 && e != 0 {
					episode := media.NewEpisode(name, s, e)
					logrus.Debugf("Adding %q for scraping", episode.UniqueTitle)
					searchItems <- episode
				}
				fmt.Println()
				fmt.Println()
			}
		}()

		// httpClient := &http.Client{Timeout: time.Second * 5}

		// Start polling providers
		// mediaProviders := []mediaprovider.Poller{
		// }
		// pollStep := pipeline.NewPollStep(repo, mediaProviders)
		// searchItems := pollStep.Poll()

		// TODO Full season torrents?
		// searchItems <- media.NewEpisode("Punisher", 2, 1)

		// Filter out items that we already have
		scrapeItems := make(chan media.Metadata)
		go func() {
			for item := range searchItems {
				m, err := repo.Fetch(item.UniqueTitle)

				if err == sql.ErrNoRows || m.Status == storage.StatusPending {
					scrapeItems <- item
					logrus.Infof("pushing %q for scraping", item.UniqueTitle)
				} else {
					logrus.Infof("skipping %q for scraping", item.UniqueTitle)
				}
			}
		}()

		// Scrape provided search items
		rarbgScraper, err := magnet.NewRarbgScraper()
		if err != nil {
			logrus.Fatalf("could not initialize rarbg: %s", err)
		}

		magnetChan := pipeline.NewScrapeStep(repo, []magnet.Scraper{rarbgScraper}).Scrape(scrapeItems)

		// Extract links from magnets
		rdExtractor := magnet.NewRealDebridExtractor(
			rd.NewRealDebrid(createToken(&config.RealDebrid), http.DefaultClient, rd.AutoRefresh),
			time.Second*10,
			false,
		)

		downloadLocations := pipeline.NewExtractStep(repo, rdExtractor, config).Extract(magnetChan)

		movieDownloader := download.NewHttpDownloader()

		downloadedItems := pipeline.NewDownloadStep(repo, movieDownloader, config.MaximumDownloadFiles).Download(downloadLocations)

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
				logrus.Debugf("Downloaded %q", item.UniqueTitle)
			}
		}()

		<-stop
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
