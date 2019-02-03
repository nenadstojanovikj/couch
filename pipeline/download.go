package pipeline

import (
	"fmt"
	"github.com/nenadstojanovikj/couch/pkg/download"
	"github.com/nenadstojanovikj/couch/pkg/media"
	"github.com/nenadstojanovikj/couch/pkg/storage"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type downloadStep struct {
	repo   *storage.MediaRepository
	getter download.Getter

	mu               sync.RWMutex
	currentDownloads map[string]interface{}

	maxDL     chan interface{}
	informers map[download.Informer]download.Informer
}

func NewDownloadStep(repo *storage.MediaRepository, getter download.Getter, maxDownloads int) *downloadStep {
	maxDL := make(chan interface{}, maxDownloads)

	for i := 0; i < maxDownloads; i++ {
		maxDL <- struct{}{}
	}

	return &downloadStep{
		repo:             repo,
		getter:           getter,
		maxDL:            maxDL,
		currentDownloads: make(map[string]interface{}),
		informers:        make(map[download.Informer]download.Informer),
	}
}

func (step *downloadStep) Download(downloads <-chan storage.Download) chan media.SearchItem {
	downloadedChan := make(chan media.SearchItem)

	go func() {
		// Start downloads
		for dl := range downloads {
			logrus.Debugf("queuing download for %q", dl.Location)
			if err := step.repo.AddDownload(dl); err != nil {
				logrus.Errorf("error while adding a download link: %s", err)
				continue
			}

			step.mu.Lock()
			if _, ok := step.currentDownloads[dl.Location]; ok {
				logrus.Debugf("skipped download for %q, already in progress", dl.Location)
				step.mu.Unlock()
				continue
			}
			step.currentDownloads[dl.Location] = nil
			step.mu.Unlock()

			// Take a token or wait until one is available
			<-step.maxDL

			logrus.Debugf("started download for %q", dl.Location)
			informer, err := step.getter.Get(dl.Item, dl.Location, dl.Destination)
			if err != nil {
				logrus.Errorf("error during download: %s", err)
				continue
			}

			if err := step.repo.UpdateDownload(informer); err != nil {
				logrus.Errorf("could not update status before download: %s", err)
			}

			step.mu.Lock()
			step.informers[informer] = informer
			step.mu.Unlock()
		}
	}()

	go func() {
		for {
			for index, informer := range step.informers {
				info := informer.Info()

				if !info.IsDone {
					continue
				}

				if info.Error == nil {
					// Publish only if there wasn't any error
					downloadedChan <- info.Item
				} else {
					logrus.Errorf("error while downloading %q: %s", info.Item.Term, info.Error)
				}

				step.maxDL <- struct{}{}
				if err := step.repo.UpdateDownload(informer); err != nil {
					logrus.Errorf("could not update status after download: %s", err)
					continue
				}

				logrus.Debugf("completed download for %q", info.Item.Term)
				delete(step.informers, index)
			}
			time.Sleep(time.Second * 5)
		}
	}()

	// Run progress
	go func() {
		infoChan := make(chan os.Signal)
		signal.Notify(infoChan, syscall.SIGUSR1)

		for {
			<-infoChan
			for _, informer := range step.informers {
				info := informer.Info()
				fmt.Printf("Progress of %s is %s\n", info.Filepath, info.ProgressBytes())
				fmt.Printf("  -> %d/%d (%.2f%%)\n", info.DownloadedBytes, info.TotalBytes, info.Progress()*100)
			}
		}
	}()

	return downloadedChan
}
