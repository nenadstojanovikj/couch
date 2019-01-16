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
func (step *scrapeStep) Scrape(searchItems <-chan media.Item) chan storage.Magnet {
	magnetChan := make(chan storage.Magnet)
	go func() {
		for item := range searchItems {
			// TODO Change check to be status instead of torrent count
			if ok, err := step.repo.HasMagnets(item.Title); err != nil {
				logrus.Errorf("error while fetching magnet info about item %s: %s", item.Title, err)
				continue
			} else if ok {
				logrus.Debugf("item %q has magnets, skipping", item.Title)
				continue
			}

			logrus.Debugf("stored %s\n", item.Title)
			err := step.repo.StoreItem(item)
			if err != nil {
				logrus.Errorf("could not store %q: %s", item.Title, err)
				continue
			}

			magnets := make([]storage.Magnet, 0)
			for _, s := range step.scrapers {
				items, err := s.Scrape(item)
				if err != nil {
					logrus.Errorf("could not scrape %s: %s", item.Title, err)
					continue
				}
				magnets = append(magnets, items...)
				logrus.Debugf("scraped %s\n", item.Title)
			}

			processors := []magnet.ProcessFunc{
				magnet.FilterQuality(storage.QualityHD, storage.Quality4K),
				magnet.FilterEncoding(storage.Encodingx264, storage.Encodingx265, storage.EncodingHEVC),
				magnet.SortQuality(true),
				magnet.SortEncoding(true),
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

			// Pushing only the first torrent
			if len(magnets) == 0 {
				logrus.Warnf("no magnets for %q", item.Title)
				continue
			}
			magnetChan <- magnets[0]
		}
	}()

	return magnetChan
}
