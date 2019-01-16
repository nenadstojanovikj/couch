package cmd

import (
	"github.com/cavaliercoder/grab"
	"github.com/nenadstojanovikj/couch/pipeline"
	"github.com/nenadstojanovikj/couch/pkg/config"
	"github.com/nenadstojanovikj/couch/pkg/download"
	"github.com/nenadstojanovikj/couch/pkg/magnet"
	"github.com/nenadstojanovikj/couch/pkg/mediaprovider"
	"github.com/nenadstojanovikj/couch/pkg/storage"
	"github.com/nenadstojanovikj/couch/pkg/web"
	"github.com/nenadstojanovikj/rd"
	"github.com/nenadstojanovikj/showrss-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"os/signal"
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

		couchWeb := web.NewWebServer(config.Port)
		go func() {
			if err := couchWeb.ListenAndServe(); err != nil {
				logrus.Fatalf("error with web server: %s", err)
			}
		}()

		// Start polling providers
		mediaProviders := []mediaprovider.Poller{
			newShowRSSProvider(config),
		}
		pollStep := pipeline.NewPollStep(repo, mediaProviders)
		searchItems := pollStep.Poll()

		// Scrape provided search items
		rarbgScraper, err := magnet.NewRarbgScraper()
		if err != nil {
			logrus.Fatalf("could not initialize rarbg: %s", err)
		}
		scrapeStep := pipeline.NewScrapeStep(repo, []magnet.Scraper{rarbgScraper})
		magnetChan := scrapeStep.Scrape(searchItems)

		// Extract links from magnets
		rdExtractor := magnet.NewRealDebridExtractor(
			rd.NewRealDebrid(createToken(&config.RealDebrid), http.DefaultClient, rd.AutoRefresh),
			time.Second*20,
			false,
		)
		extractStep := pipeline.NewExtractStep(repo, rdExtractor, config)
		locationChannel := extractStep.Extract(magnetChan)

		// Download final links
		// TODO Use couch http client
		grabber := grab.NewClient()
		movieDownloader := download.NewHttpDownloader(grabber)
		downloadStep := pipeline.NewDownloadStep(repo, movieDownloader)
		downloadedItem := downloadStep.Download(locationChannel)

		for item := range downloadedItem {
			logrus.Printf("Downloaded %q", item.Title)
		}

		<-stop
	}
}

func newShowRSSProvider(config *config.Config) *mediaprovider.ShowRSSProvider {
	return mediaprovider.NewShowRSSProvider(
		time.Second*10,
		config.ShowRss.PersonalFeed,
		showrss.NewClient(http.DefaultClient),
	)
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
