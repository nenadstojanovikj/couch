package download

import (
	"github.com/cavaliercoder/grab"
)

type HttpDownloader struct {
	grab *grab.Client
}

type GrabFile struct {
	response *grab.Response
}

func NewHttpDownloader(grab *grab.Client) *HttpDownloader {
	return &HttpDownloader{
		grab: grab,
	}
}

func (f *GrabFile) Info() *Info {
	var err error
	if f.response.IsComplete() {
		err = f.response.Err()
	}

	return &Info{
		IsDone:          f.response.IsComplete(),
		Error:           err,
		Filepath:        f.response.Filename,
		TotalBytes:      f.response.Size,
		DownloadedBytes: f.response.BytesComplete(),
	}
}

func (d *HttpDownloader) Get(url string, destination string) (Informer, error) {
	req, err := grab.NewRequest(destination, url)
	if err != nil {
		return nil, err
	}

	return &GrabFile{response: d.grab.Do(req)}, nil
}
