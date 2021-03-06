package pipeline

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nenad/couch/pkg/download"
	"github.com/nenad/couch/pkg/media"
	"github.com/nenad/couch/pkg/notifications"
	"github.com/nenad/couch/pkg/storage"
	"github.com/sirupsen/logrus"
)

// DownloadStep is the final step in the pipeline
type DownloadStep struct {
	repo   *storage.MediaRepository
	getter download.Getter

	mu               sync.RWMutex
	currentDownloads map[string]interface{}

	maxDL     chan struct{}
	informers map[download.Informer]download.Informer
	notifier  notifications.Notifier
}

// NewDownloadStep returns a pipeline step to download items
// TODO handle cases where there is a retry loop of downloads in "error" state
func NewDownloadStep(repo *storage.MediaRepository, getter download.Getter, maxDownloads int, notifier notifications.Notifier) *DownloadStep {
	maxDL := make(chan struct{}, maxDownloads)

	return &DownloadStep{
		repo:             repo,
		getter:           getter,
		maxDL:            maxDL,
		currentDownloads: make(map[string]interface{}),
		informers:        make(map[download.Informer]download.Informer),
		notifier:         notifier,
	}
}

func (step *DownloadStep) Download(downloads <-chan storage.Download) chan media.SearchItem {
	downloadedChan := make(chan media.SearchItem)

	go func() {
		// Start downloads
		for dl := range downloads {
			logrus.Debugf("queueing download for %q", dl.Remote)
			if err := step.repo.AddDownload(dl); err != nil {
				logrus.Errorf("error while adding a download link: %s", err)
				continue
			}
			step.notifier.OnQueued(dl.Item)

			step.mu.Lock()
			if _, ok := step.currentDownloads[dl.Remote]; ok {
				logrus.Debugf("skipped download for %q, already in progress", dl.Remote)
				step.mu.Unlock()
				continue
			}
			step.currentDownloads[dl.Remote] = nil
			step.mu.Unlock()

			// Acquire a token or wait until one is available
			step.maxDL <- struct{}{}

			logrus.Debugf("started download for %q", dl.Remote)
			informer, err := step.getter.Get(dl.Item, dl.Remote, dl.Local)
			if err != nil {
				logrus.Errorf("error during download: %s", err)
				continue
			}

			info := informer.Info()

			if err := step.repo.UpdateDownload(dl.Item.Term, info.Filepath, info.IsDone, info.Error); err != nil {
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

				// Release once done
				<-step.maxDL
				if err := step.repo.UpdateDownload(info.Item.Term, info.Url, info.IsDone, info.Error); err != nil {
					logrus.Errorf("could not update status after download: %s", err)
					continue
				}

				logrus.Debugf("completed download for %q", info.Url)
				step.notifier.OnFinish(info.Item)
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
