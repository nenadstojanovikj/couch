package magnet

import (
	"github.com/nenadstojanovikj/couch/pkg/storage"
)

var allowedExtensions = map[string]bool{".mp4": true, ".mkv": true, ".avi": true, ".m4v": true, ".wmv": true}

type Extractor interface {
	Extract(magnet storage.Magnet) ([]string, error)
}
