package download

import (
	"fmt"
	"github.com/nenadstojanovikj/couch/pkg/media"
)

const (
	TypeTorrent = "torrent"
	TypeHTTP    = "http"
)

type Info struct {
	Item            media.SearchItem
	Filepath        string
	TotalBytes      int64
	DownloadedBytes int64
	IsDone          bool
	Error           error
	Url             string
}

// Progress returns the progress of the file from 0 to 1
func (f *Info) Progress() float64 {
	return float64(f.DownloadedBytes) / float64(f.TotalBytes)
}

// Progress returns the progress of the file from 0 to 1
func (f *Info) ProgressBytes() string {
	return byteCountDecimal(f.DownloadedBytes) + " / " + byteCountDecimal(f.TotalBytes)
}

type Getter interface {
	Get(item media.SearchItem, url string, destination string) (Informer, error)
}

type Informer interface {
	Info() *Info
}

func byteCountDecimal(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}
