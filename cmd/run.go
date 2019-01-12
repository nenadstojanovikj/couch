package cmd

import (
	"fmt"
	"github.com/nenadstojanovikj/couch/pkg/config"
	"github.com/nenadstojanovikj/couch/pkg/magnet"
	"github.com/nenadstojanovikj/couch/pkg/media"
	"github.com/nenadstojanovikj/couch/pkg/mediaprovider"
	"github.com/nenadstojanovikj/couch/pkg/storage"
	"github.com/nenadstojanovikj/couch/pkg/web"
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

func run(config *config.Config, repo *storage.MediaRepository) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		stop := make(chan os.Signal)
		signal.Notify(stop, os.Interrupt, os.Kill, syscall.SIGTERM)

		couchWeb := web.NewWebServer(config.Port)
		go func() {
			if err := couchWeb.ListenAndServe(); err != nil {
				logrus.Fatalf("error with web server: %s", err)
			}
		}()

		downloadItems := make(chan media.Item, 10)
		// mediaProviders := []mediaprovider.Poller{
		// 	newShowRSSProvider(config),
		// }
		//
		// downloadItems := make(chan media.Item, 10)
		// for _, provider := range mediaProviders {
		// 	go func(provider mediaprovider.Poller) {
		// 		for {
		// 			items, err := provider.Poll()
		// 			if err != nil {
		// 				continue
		// 			}
		//
		// 			for _, item := range items {
		// 				downloadItems <- item
		// 			}
		//
		// 			time.Sleep(provider.Timeout())
		// 		}
		// 	}(provider)
		// }

		downloadItems <- media.NewMovie("The Dark Knight", 2008, nil)

		rarbgScraper, err := magnet.NewRarbgScraper()
		if err != nil {
			logrus.Fatalf("could not initialize rarbg: %s", err)
		}

		scrapers := []magnet.Scraper{rarbgScraper}

		go func() {
			for item := range downloadItems {
				fmt.Printf("Stored %s\n", item.Title)
				err := repo.StoreItem(item)
				if err != nil {
					fmt.Println(err)
				}

				for _, s := range scrapers {
					fmt.Printf("Scraped %s\n", item.Title)
					s.Scrape(item)
				}
			}
		}()

		// // Trakt TV Provider
		// err, items := trakt.GetListItems(config.TraktDownloadList)
		// for _, item := range items {
		// 	switch item.Type {
		// 	case "movie":
		// 		downloadItems <- model.NewMovie(item.Title, item.Year)
		// 	case "episode":
		// 		downloadItems <- model.NewEpisode(item.Title, item.Season, item.Episode)
		// 	default:
		// 		break
		// 	}
		// }
		// time.Sleep(time.Second * config.TraktPoll)
		// trakt.Poll(60, )

		// Wait until it's killed by user
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
