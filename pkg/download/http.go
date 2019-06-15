package download

import (
	"github.com/cavaliercoder/grab"
	"github.com/nenad/couch/pkg/media"
)

type HttpDownloader struct {
	grab *grab.Client
}

type grabFile struct {
	response *grab.Response
	item     media.SearchItem
}

func NewHttpDownloader() *HttpDownloader {
	return &HttpDownloader{
		grab: grab.NewClient(),
	}
}

func (f *grabFile) Info() *Info {
	var err error
	if f.response.IsComplete() {
		err = f.response.Err()
	}

	return &Info{
		Item:            f.item,
		IsDone:          f.response.IsComplete(),
		Error:           err,
		Filepath:        f.response.Filename,
		TotalBytes:      f.response.Size,
		DownloadedBytes: f.response.BytesComplete(),
		Url:             f.response.Request.URL().String(),
	}
}

func (d *HttpDownloader) Get(item media.SearchItem, url string, destination string) (Informer, error) {
	req, err := grab.NewRequest(destination, url)
	if err != nil {
		return nil, err
	}

	return &grabFile{response: d.grab.Do(req), item: item}, nil
}
