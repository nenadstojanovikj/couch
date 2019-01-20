package pipeline

import (
	"fmt"
	"github.com/nenadstojanovikj/couch/pkg/download"
	"github.com/nenadstojanovikj/couch/pkg/media"
	"github.com/nenadstojanovikj/couch/pkg/storage"
	"github.com/sirupsen/logrus"
	"time"
)

type downloadStep struct {
	repo   *storage.MediaRepository
	getter download.Getter

	currentDownloads map[string]interface{}

	maxDL chan interface{}
}

func NewDownloadStep(repo *storage.MediaRepository, getter download.Getter, maxDownloads int) *downloadStep {
	maxDL := make(chan interface{}, maxDownloads)

	for i := 0; i < maxDownloads; i++ {
		maxDL <- nil
	}

	return &downloadStep{
		repo:             repo,
		getter:           getter,
		maxDL:            maxDL,
		currentDownloads: make(map[string]interface{}),
	}
}

func (step *downloadStep) Download(downloads <-chan storage.Download) chan media.Item {
	downloadedChan := make(chan media.Item)

	go func() {
		for dl := range downloads {
			logrus.Debugf("downloading %q", dl.Location)
			if _, ok := step.currentDownloads[dl.Location]; ok {
				continue
			}
			step.currentDownloads[dl.Location] = nil

			<-step.maxDL

			check, err := step.getter.Get(dl.Location, dl.Destination)
			if err != nil {
				logrus.Errorf("error during download: %s", err)
				continue
			}

			if err := step.repo.Status(dl.Item.Title, storage.StatusDownloading); err != nil {
				logrus.Errorf("could not update status after download: %s", err)
			}

			// Run progress
			go func(dl storage.Download, informer download.Informer) {
				for {
					if informer.Info().TotalBytes == informer.Info().DownloadedBytes {
						downloadedChan <- dl.Item
						if err := step.repo.Status(dl.Item.Title, storage.StatusDownloaded); err != nil {
							logrus.Errorf("could not update status after download: %s", err)
						}

						break
					}
					if informer.Info().Error != nil {
						logrus.Errorf("error while downloading: %s", informer.Info().Error)
						break
					}
					fmt.Printf("\rProgress of %s is %f", dl.Location, informer.Info().Progress())
					time.Sleep(time.Second)
				}
			}(dl, check)
		}
	}()

	return downloadedChan
}
