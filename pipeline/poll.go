package pipeline

import (
	"database/sql"
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
	searches := make(chan media.SearchItem, 10)

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
					searches <- item
				}

				time.Sleep(provider.Interval())
			}
		}(provider)
	}

	newSearches := make(chan media.SearchItem, 10)
	go func() {
		for item := range searches {
			m, err := step.repo.Fetch(item.Term)

			if m.Status == storage.StatusPending {
				newSearches <- item
				continue
			}

			if err == sql.ErrNoRows {
				err := step.repo.StoreItem(item)
				if err != nil {
					logrus.Errorf("could not store %q: %s", item.Term, err)
					continue
				}

				newSearches <- item
				logrus.Infof("pushing %q for scraping", item.Term)
				continue
			}

			logrus.Infof("skipping %q for scraping, already in database", item.Term)
		}
	}()

	return newSearches
}
