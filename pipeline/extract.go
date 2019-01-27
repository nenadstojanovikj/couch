package pipeline

import (
	"fmt"
	"github.com/nenadstojanovikj/couch/pkg/config"
	"github.com/nenadstojanovikj/couch/pkg/magnet"
	"github.com/nenadstojanovikj/couch/pkg/media"
	"github.com/nenadstojanovikj/couch/pkg/storage"
	"github.com/sirupsen/logrus"
	"path"
	"regexp"
	"strconv"
	"sync"
)

type (
	extractStep struct {
		repo      *storage.MediaRepository
		extractor magnet.Extractor
		config    *config.Config

		mu              sync.RWMutex
		currentExtracts map[media.Title]interface{}
	}
)

func NewExtractStep(repo *storage.MediaRepository, extractor magnet.Extractor, config *config.Config) *extractStep {
	return &extractStep{
		repo:            repo,
		extractor:       extractor,
		config:          config,
		currentExtracts: make(map[media.Title]interface{}),
	}
}

var tvShowRegex = regexp.MustCompile("(.*) S([0-9]{2})E([0-9]{2})")

func (step *extractStep) Extract(magnetChan chan storage.Magnet) chan storage.Download {
	dlMap := make(chan storage.Download, 10)
	go func() {
		for mag := range magnetChan {
			step.mu.RLock()
			if _, ok := step.currentExtracts[mag.Item.Title]; ok {
				step.mu.RUnlock()
				logrus.Debugf("skipped %q", mag.Item.Title)
				continue
			}
			step.mu.RUnlock()

			step.mu.Lock()
			step.currentExtracts[mag.Item.Title] = nil
			step.mu.Unlock()

			go func(m storage.Magnet) {
				logrus.Debugf("extracting %q", m.Item.Title)
				if err := step.repo.Status(m.Item.Title, storage.StatusExtracting); err != nil {
					logrus.Errorf("could not update status after download: %s", err)
				}

				urls, err := step.extractor.Extract(m.Location)
				if err != nil {
					logrus.Errorf("could not extract link %s: %s", m.Location, err)
					return
				}

				for _, url := range urls {
					dlLocation := storage.Download{
						Location: url,
					}

					// TODO Move to download package
					// Function to generate path
					switch m.Item.Type {
					case media.TypeMovie:
						// Decide if file needs to be renamed
						dlLocation.Destination = path.Join(step.config.MoviesPath, path.Base(url))
					case media.TypeEpisode:
						matches := tvShowRegex.FindAllStringSubmatch(string(m.Item.Title), -1)
						name := matches[0][1]
						season, _ := strconv.Atoi(matches[0][2])

						// TODO Check if TV Shows are in separate dirs
						dlLocation.Destination = path.Join(step.config.TVShowsPath, fmt.Sprintf("%s/Season %02d/%s", name, season, path.Base(url)))
					}
					dlLocation.Item = m.Item

					if err := step.repo.AddDownload(dlLocation); err != nil {
						logrus.Errorf("could not add download for %q: %s", dlLocation.Item.Title, err)
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
