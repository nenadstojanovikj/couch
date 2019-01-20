package refresh

import (
	"github.com/nenadstojanovikj/couch/pkg/storage"
)

func Download(repo *storage.MediaRepository, toDownload chan storage.Download) error {
	downloads, err := repo.NonStartedDownloads()
	if err != nil {
		return err
	}

	for _, t := range downloads {
		toDownload <- t
	}

	return nil
}
