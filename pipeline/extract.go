package pipeline

import (
	"github.com/nenadstojanovikj/couch/pkg/config"
	"github.com/nenadstojanovikj/couch/pkg/magnet"
	"github.com/nenadstojanovikj/couch/pkg/media"
	"github.com/nenadstojanovikj/couch/pkg/storage"
	"github.com/sirupsen/logrus"
	"path"
)

type (
	extractStep struct {
		repo      *storage.MediaRepository
		extractor magnet.Extractor
		config    *config.Config
	}

	downloadLocation struct {
		Location    string
		Destination string
	}
)

func NewExtractStep(repo *storage.MediaRepository, extractor magnet.Extractor, config *config.Config) *extractStep {
	return &extractStep{
		repo:      repo,
		extractor: extractor,
		config:    config,
	}
}

func (step *extractStep) Extract(magnetChan chan storage.Magnet) chan downloadLocation {
	dlMap := make(chan downloadLocation, 10)
	// TODO Pick between rdExtractor and localExtractor
	go func() {
		for mag := range magnetChan {
			go func(m storage.Magnet) {
				url, err := step.extractor.Extract(m.Location)
				if err != nil {
					logrus.Errorf("could not extract link %s: %s", m.Location, err)
					return
				}

				if err := step.repo.AddLinks(m.Item.Title, []string{url}); err != nil {
					logrus.Errorf("could not add link %s for %q: %s", url, m.Item.Title, err)
					return
				}

				dlLocation := downloadLocation{
					Location: url,
				}

				// TODO Move to download package
				// Function to generate path
				switch m.Item.Type {
				case media.TypeMovie:
					// Decide if file needs to be renamed
					dlLocation.Destination = path.Join(step.config.MoviesPath, path.Base(url))
				case media.TypeEpisode:
					// Check if TV Shows are in separate dirs
					dlLocation.Destination = path.Join(step.config.TVShowsPath, string(m.Item.Title))
				}

				dlMap <- dlLocation
			}(mag)
		}
		close(dlMap)
	}()
	return dlMap
}
