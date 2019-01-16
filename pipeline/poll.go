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

func (step *pollStep) Poll() chan media.Item {
	searchItems := make(chan media.Item, 10)

	for _, provider := range step.pollers {
		// TODO Add pauseChan which would stop the polling for a specified provider
		go func(provider mediaprovider.Poller) {
			for {
				items, err := provider.Poll()
				if err != nil {
					logrus.Errorf("could not poll %t: %s", provider, err)
					continue
				}

				for _, item := range items {
					searchItems <- item
				}

				time.Sleep(provider.Interval())
			}
		}(provider)
	}

	return searchItems
}
