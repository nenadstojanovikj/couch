package pipeline

import (
	"github.com/nenadstojanovikj/couch/pkg/media"
	"github.com/nenadstojanovikj/couch/pkg/mediaprovider"
	"github.com/nenadstojanovikj/couch/pkg/storage"
	"github.com/sirupsen/logrus"
	"time"
)

type pollStep struct {
	pollers []mediaprovider.Poller
	repo    *storage.MediaRepository
}

func NewPollStep(repo *storage.MediaRepository, pollers []mediaprovider.Poller) *pollStep {
	return &pollStep{
		repo:    repo,
		pollers: pollers,
	}
}

func (step *pollStep) Poll() chan media.SearchItem {
	searchItems := make(chan media.SearchItem, 10)

	for _, provider := range step.pollers {
		// TODO Add pauseChan which would stop the polling for a specified provider
		go func(provider mediaprovider.Poller) {
			for {
				items, err := provider.Poll()
				if err != nil {
					logrus.Errorf("could not poll %T: %s", provider, err)
				}

				for _, item := range items {
					logrus.Debugf("fetched %q for searching", item.Term)
					searchItems <- item
				}

				time.Sleep(provider.Interval())
			}
		}(provider)
	}

	return searchItems
}
