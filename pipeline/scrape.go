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
func (step *scrapeStep) Scrape(searchItems <-chan media.SearchItem) chan storage.Magnet {
	magnetChan := make(chan storage.Magnet)
	go func() {
		for item := range searchItems {
			logrus.Debugf("scraping %q", item.Term)

			var magnets []storage.Magnet
			for _, s := range step.scrapers {
				items, err := s.Scrape(item)
				if err != nil {
					logrus.Errorf("could not scrape %q: %s", item.Term, err)
					continue
				}
				magnets = append(magnets, items...)
				logrus.Debugf("scraped %q", item.Term)
			}

			// TODO Store Seeders in magnet?
			processors := []magnet.ProcessFunc{
				magnet.FilterQuality(storage.QualityFHD, storage.QualityFHD),
				magnet.FilterEncoding(storage.Encodingx264, storage.Encodingx265),
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
				logrus.Warnf("no magnets for %q", item.Term)
				continue
			}

			if err := step.repo.Status(item.Term, storage.StatusScraped); err != nil {
				logrus.Errorf("error while updating status in database: %s", err)
				continue
			}

			// Pushing only the best torrent
			magnetChan <- magnets[0]
			logrus.Debugf("pushed magnet %q for download", magnets[0].Location)
		}
	}()

	return magnetChan
}
