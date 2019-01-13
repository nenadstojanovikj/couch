package cmd

import (
	"fmt"
	"github.com/cavaliercoder/grab"
	"github.com/nenadstojanovikj/couch/pkg/config"
	"github.com/nenadstojanovikj/couch/pkg/download"
	"github.com/nenadstojanovikj/couch/pkg/magnet"
	"github.com/nenadstojanovikj/couch/pkg/media"
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
	"path"
	"syscall"
	"time"
)

type downloadLocation struct {
	location    string
	destination string
}

func NewAppCommand(config *config.Config, repo *storage.MediaRepository) *cobra.Command {
	return &cobra.Command{
		Use: "run",
		Run: run(config, repo),
	}
}

// Poll search providers -> search_items
// search_items -> Scrape torrents -> magnet_info
// magnet_info -> Extract links & files -> torrent_files & realdebrid
// torrent_files & realdebrid -> Download

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

		searchItems := make(chan media.Item, 10)

		// searchItems := providers.Poll()

		// go func() {
		// 	for item := range searchItems {
		//
		// 	}
		// }()

		// mediaProviders := []mediaprovider.Poller{
		// 	newShowRSSProvider(config),
		// }
		//
		// searchItems := make(chan media.Item, 10)
		// for _, provider := range mediaProviders {
		// 	go func(provider mediaprovider.Poller) {
		// 		for {
		// 			items, err := provider.Poll()
		// 			if err != nil {
		// 				continue
		// 			}
		//
		// 			for _, item := range items {
		// 				searchItems <- item
		// 			}
		//
		// 			time.Sleep(provider.Interval())
		// 		}
		// 	}(provider)
		// }

		searchItems <- media.NewMovie("Venom", 2018, nil)

		rarbgScraper, err := magnet.NewRarbgScraper()
		if err != nil {
			logrus.Fatalf("could not initialize rarbg: %s", err)
		}

		scrapers := []magnet.Scraper{rarbgScraper}

		rdToken := createToken(&config.RealDebrid)

		rdExtractor := magnet.NewRealDebridExtractor(
			rd.NewRealDebrid(rdToken, http.DefaultClient, rd.AutoRefresh),
			time.Second*20,
			false,
		)

		var extractor magnet.Extractor = rdExtractor

		dlMap := make(chan downloadLocation, 10)

		magnetChan := make(chan storage.Magnet)

		go func() {
			for item := range searchItems {
				// TODO Change check to be status instead of torrent count
				if ok, err := repo.HasMagnets(item.Title); err != nil {
					logrus.Errorf("error while fetching magnet info about item %s: %s", item.Title, err)
					continue
				} else if ok {
					logrus.Debugf("item %q has magnets, skipping", item.Title)
					continue
				}

				logrus.Debugf("stored %s\n", item.Title)
				err := repo.StoreItem(item)
				if err != nil {
					logrus.Errorf("could not store %q: %s", item.Title, err)
					continue
				}

				magnets := make([]storage.Magnet, 0)
				for _, s := range scrapers {
					items, err := s.Scrape(item)
					if err != nil {
						logrus.Errorf("could not scrape %s: %s", item.Title, err)
						continue
					}
					magnets = append(magnets, items...)
					logrus.Debugf("scraped %s\n", item.Title)
				}

				filters := []magnet.Filter{
					magnet.FilterQuality(storage.QualityHD, storage.Quality4K),
					magnet.FilterEncoding(storage.Encodingx264, storage.Encodingx265, storage.EncodingHEVC),
				}

				sorters := []magnet.Sort{
					magnet.SortQuality(true),
					magnet.SortEncoding(true),
				}

				// Filter and sort
				for _, f := range filters {
					magnets = f(magnets)
				}
				for _, s := range sorters {
					s(magnets)
				}

				for rating, m := range magnets {
					m.Rating = rating
					if err := repo.AddTorrent(m); err != nil {
						logrus.Errorf("could not add magnet %s: %s", m.Location, err)
						continue
					}
				}

				// Pushing only the first torrent
				magnetChan <- magnets[0]
			}
		}()

		go func() {
			for mag := range magnetChan {
				go func(m storage.Magnet) {
					url, err := extractor.Extract(m.Location)
					if err != nil {
						logrus.Errorf("could not extract link %s: %s", m.Location, err)
						return
					}

					if err := repo.AddLinks(m.Item.Title, []string{url}); err != nil {
						logrus.Errorf("could not add link %s for %q: %s", url, m.Item.Title, err)
						return
					}

					dlLocation := downloadLocation{
						location: url,
					}

					// TODO Move to download package
					// Function to generate path
					switch m.Item.Type {
					case media.TypeMovie:
						// Decide if file needs to be renamed
						dlLocation.destination = path.Join(config.MoviesPath, path.Base(url))
					case media.TypeEpisode:
						// Check if TV Shows are in separate dirs
						dlLocation.destination = path.Join(config.TVShowsPath, string(m.Item.Title))
					}

					dlMap <- dlLocation
				}(mag)
			}
		}()

		// TODO Use couch http client
		grabber := grab.NewClient()
		movieDownloader := download.NewHttpDownloader(grabber)
		go func() {
			for dl := range dlMap {
				check, err := movieDownloader.Get(dl.location, dl.destination)
				if err != nil {
					logrus.Errorf("error during download: %s", err)
					continue
				}

				// Run progress
				go func(dl downloadLocation, informer download.Informer) {
					for {
						fmt.Printf("\rProgress of %s is %f", dl.location, informer.Info().Progress())
						time.Sleep(time.Second)
						if informer.Info().TotalBytes == informer.Info().DownloadedBytes {
							break
						}
					}
				}(dl, check)
			}
		}()

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
