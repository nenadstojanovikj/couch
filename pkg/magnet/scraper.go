package magnet

import (
	"github.com/nenadstojanovikj/couch/pkg/media"
	"github.com/nenadstojanovikj/couch/pkg/storage"
)

type Scraper interface {
	Scrape(item media.Metadata) ([]storage.Magnet, error)
}
