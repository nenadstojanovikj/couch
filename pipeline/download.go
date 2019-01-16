package pipeline

import (
	"github.com/nenadstojanovikj/couch/pkg/download"
	"github.com/nenadstojanovikj/couch/pkg/media"
	"github.com/nenadstojanovikj/couch/pkg/storage"
	"github.com/sirupsen/logrus"
	"time"
)

type downloadStep struct {
	repo   *storage.MediaRepository
	getter download.Getter
}

func NewDownloadStep(repo *storage.MediaRepository, getter download.Getter) *downloadStep {
	return &downloadStep{
		repo:   repo,
		getter: getter,
	}
}

// TODO Take processors from constructor/config
func (step *downloadStep) Download(dlMap <-chan downloadLocation) chan media.Item {
	downloadedChan := make(chan media.Item)
	// TODO Use couch http client
	// TODO Use given downloader
	// grabber := grab.NewClient()
	// movieDownloader := download.NewHttpDownloader(grabber)
	go func() {
		for dl := range dlMap {
			check, err := step.getter.Get(dl.Location, dl.Destination)
			if err != nil {
				logrus.Errorf("error during download: %s", err)
				continue
			}

			// Run progress
			go func(dl downloadLocation, informer download.Informer) {
				for {
					if informer.Info().TotalBytes == informer.Info().DownloadedBytes {
						downloadedChan <- media.Item{Title: "HELLO", Type: media.TypeMovie}

						break
					}
					// fmt.Printf("\rProgress of %s is %f", dl.Location, informer.Info().Progress())
					time.Sleep(time.Second)
					// if informer.Info().TotalBytes == informer.Info().DownloadedBytes {
					// 	break
					// }
				}
			}(dl, check)
		}
	}()
	return downloadedChan
}
