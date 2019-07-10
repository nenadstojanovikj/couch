package pipeline

import (
	"sync"

	"github.com/nenad/couch/pkg/config"
	"github.com/nenad/couch/pkg/magnet"
	"github.com/nenad/couch/pkg/media"
	"github.com/nenad/couch/pkg/storage"
	"github.com/sirupsen/logrus"
)

type (
	extractStep struct {
		repo      *storage.MediaRepository
		extractor magnet.Extractor
		config    config.Config

		mu              sync.RWMutex
		currentExtracts map[string]interface{}
	}
)

func NewExtractStep(repo *storage.MediaRepository, extractor magnet.Extractor, config config.Config) *extractStep {
	return &extractStep{
		repo:            repo,
		extractor:       extractor,
		config:          config,
		currentExtracts: make(map[string]interface{}),
	}
}

func (step *extractStep) Extract(magnetChan chan storage.Magnet) chan storage.Download {
	dlMap := make(chan storage.Download, 10)
	go func() {
		for mag := range magnetChan {
			step.mu.Lock()
			if _, ok := step.currentExtracts[mag.Item.Term]; ok {
				step.mu.Unlock()
				logrus.Debugf("skipped extract of %q as it is in progress", mag.Item.Term)
				continue
			}
			step.currentExtracts[mag.Item.Term] = nil
			step.mu.Unlock()

			go func(m storage.Magnet) {
				logrus.Debugf("extracting %q", m.Item.Term)
				if err := step.repo.Status(m.Item.Term, storage.StatusExtracting); err != nil {
					logrus.Errorf("could not update status after download: %s", err)
				}

				urls, err := step.extractor.Extract(m)
				if err != nil {
					logrus.Errorf("could not extract link %s: %s", m.Location, err)
					step.mu.Lock()
					delete(step.currentExtracts, m.Item.Term)
					step.mu.Unlock()
					return
				}

				for _, url := range urls {
					var dest string
					switch m.Item.Type {
					case media.TypeMovie:
						dest = m.Item.Path(step.config.MoviesPath, url)
					case media.TypeEpisode, media.TypeSeason:
						dest = m.Item.Path(step.config.TVShowsPath, url)
					}

					dlLocation := storage.Download{
						Location:    url,
						Destination: dest,
						Item:        m.Item,
					}

					if err := step.repo.AddDownload(dlLocation); err != nil {
						logrus.Errorf("could not add download for %q: %s", dlLocation.Item.Term, err)
						continue
					}

					dlMap <- dlLocation
				}
			}(mag)
		}
		close(dlMap)
	}()
	return dlMap
}
