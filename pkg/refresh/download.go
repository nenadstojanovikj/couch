package refresh

import (
	"github.com/nenad/couch/pkg/storage"
)

func Download(repo *storage.MediaRepository, toDownload chan storage.Download) error {
	downloads, err := repo.InProgressDownloads()
	if err != nil {
		return err
	}

	for _, t := range downloads {
		toDownload <- t
	}

	return nil
}
