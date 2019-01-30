package pipeline

import (
	"github.com/nenadstojanovikj/couch/pkg/magnet"
	"github.com/nenadstojanovikj/couch/pkg/media"
	"github.com/nenadstojanovikj/couch/pkg/storage"
	"github.com/sirupsen/logrus"
)

type scrapeStep struct {
	repo     *storage.MediaRepository
	scrapers []magnet.Scraper
}

func NewScrapeStep(repo *storage.MediaRepository, scrapers []magnet.Scraper) *scrapeStep {
	return &scrapeStep{
		repo:     repo,
		scrapers: scrapers,
	}
}

// TODO Take processors from constructor/config
func (step *scrapeStep) Scrape(searchItems <-chan media.Metadata) chan storage.Magnet {
	magnetChan := make(chan storage.Magnet)
	go func() {
		for item := range searchItems {
			logrus.Debugf("scraping %q", item.UniqueTitle)
			err := step.repo.StoreItem(item)
			if err != nil {
				logrus.Errorf("could not store %q: %s", item.UniqueTitle, err)
				continue
			}

			var magnets []storage.Magnet
			for _, s := range step.scrapers {
				items, err := s.Scrape(item)
				if err != nil {
					logrus.Errorf("could not scrape %q: %s", item.UniqueTitle, err)
					continue
				}
				magnets = append(magnets, items...)
				logrus.Debugf("scraped %q", item.UniqueTitle)
			}

			// TODO Seeders?
			processors := []magnet.ProcessFunc{
				magnet.FilterQuality(storage.QualityHD, storage.Quality4K),
				magnet.FilterEncoding(storage.Encodingx264, storage.Encodingx265, storage.EncodingHEVC),
				magnet.SortSize(false),
				magnet.SortEncoding(true),
				magnet.SortQuality(true),
			}

			// Filter and sort
			for _, f := range processors {
				magnets = f(magnets)
			}

			for rating, m := range magnets {
				m.Rating = rating
				if err := step.repo.AddTorrent(m); err != nil {
					logrus.Errorf("could not add magnet %s: %s", m.Location, err)
					continue
				}
			}

			if len(magnets) == 0 {
				logrus.Warnf("no magnets for %q", item.UniqueTitle)
				continue
			}

			if err := step.repo.Status(item.UniqueTitle, storage.StatusScraped); err != nil {
				logrus.Errorf("error while updating status in database: %s", err)
				continue
			}

			// Pushing only the best torrent
			magnetChan <- magnets[0]
		}
	}()

	return magnetChan
}
