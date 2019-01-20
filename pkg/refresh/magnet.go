package refresh

import (
	"github.com/nenadstojanovikj/couch/pkg/storage"
)

func Magnet(repo *storage.MediaRepository, toScrape chan storage.Magnet) error {
	torrents, err := repo.NonStartedTorrents()
	if err != nil {
		return err
	}

	for _, t := range torrents {
		toScrape <- t
	}

	return nil
}
