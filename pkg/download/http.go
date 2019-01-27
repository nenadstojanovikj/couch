package download

import (
	"github.com/cavaliercoder/grab"
	"github.com/nenadstojanovikj/couch/pkg/media"
)

type HttpDownloader struct {
	grab *grab.Client
}

type GrabFile struct {
	response *grab.Response
	item     media.Item
}

func NewHttpDownloader() *HttpDownloader {
	return &HttpDownloader{
		grab: grab.NewClient(),
	}
}

func (f *GrabFile) Info() *Info {
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
	}
}

func (d *HttpDownloader) Get(item media.Item, url string, destination string) (Informer, error) {
	req, err := grab.NewRequest(destination, url)
	if err != nil {
		return nil, err
	}

	return &GrabFile{response: d.grab.Do(req), item: item}, nil
}
