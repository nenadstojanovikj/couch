package magnet

import (
	"github.com/nenad/couch/pkg/media"
	"github.com/nenad/couch/pkg/storage"
)

type Scraper interface {
	Scrape(item media.SearchItem) ([]storage.Magnet, error)
}
