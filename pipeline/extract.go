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
		currentExtracts map[string]interface{}
	}
)

func NewExtractStep(repo *storage.MediaRepository, extractor magnet.Extractor, config *config.Config) *extractStep {
	return &extractStep{
		repo:            repo,
		extractor:       extractor,
		config:          config,
		currentExtracts: make(map[string]interface{}),
	}
}

var tvShowRegex = regexp.MustCompile("(.*) S([0-9]{2})")

func (step *extractStep) Extract(magnetChan chan storage.Magnet) chan storage.Download {
	dlMap := make(chan storage.Download, 10)
	go func() {
		for mag := range magnetChan {
			step.mu.Lock()
			if _, ok := step.currentExtracts[mag.Item.UniqueTitle]; ok {
				step.mu.Unlock()
				logrus.Debugf("skipped extract of %q as it is in progress", mag.Item.UniqueTitle)
				continue
			}
			step.currentExtracts[mag.Item.UniqueTitle] = nil
			step.mu.Unlock()

			go func(m storage.Magnet) {
				logrus.Debugf("extracting %q", m.Item.UniqueTitle)
				if err := step.repo.Status(m.Item.UniqueTitle, storage.StatusExtracting); err != nil {
					logrus.Errorf("could not update status after download: %s", err)
				}

				urls, err := step.extractor.Extract(m)
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
					case media.TypeSeason:
						matches := tvShowRegex.FindAllStringSubmatch(string(m.Item.UniqueTitle), -1)
						name := matches[0][1]
						season, _ := strconv.Atoi(matches[0][2])

						// TODO Check if TV Shows are in separate dirs
						dlLocation.Destination = path.Join(step.config.TVShowsPath, fmt.Sprintf("%s/Season %d/%s", name, season, path.Base(url)))
					}
					dlLocation.Item = m.Item

					if err := step.repo.AddDownload(dlLocation); err != nil {
						logrus.Errorf("could not add download for %q: %s", dlLocation.Item.UniqueTitle, err)
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
