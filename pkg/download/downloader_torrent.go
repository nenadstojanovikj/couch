package download

import (
	"fmt"
	"github.com/anacrolix/torrent"
	torStorage "github.com/anacrolix/torrent/storage"
	"github.com/nenad/couch/pkg/config"
	"github.com/nenad/couch/pkg/media"
	"github.com/nenad/couch/pkg/storage"
	"path"
)

type torrentDownloader struct {
	repo   *storage.MediaRepository
	config *config.Config
}

func NewTorrentDownloader(repo *storage.MediaRepository, config *config.Config) *torrentDownloader {
	return &torrentDownloader{
		repo:   repo,
		config: config,
	}
}

type torrentStatus struct {
	file     *torrent.File
	item     media.SearchItem
	filepath string
}

func (s *torrentStatus) Info() *Info {
	var err error
	if s.file.Torrent().Stats().TotalPeers == 0 {
		err = fmt.Errorf("torrent for %q has no seeders", s.item.Term)
	}

	done := false
	if err != nil {
		done = true
	}

	pieces := s.file.State()

	var total, completed int64
	for _, p := range pieces {
		total += p.Bytes

		if p.Complete {
			completed += p.Bytes
		}
	}

	if total == completed {
		done = true
	}

	return &Info{
		Url:             s.file.Path(),
		Error:           err,
		IsDone:          done,
		TotalBytes:      total,
		DownloadedBytes: completed,
		Filepath:        s.filepath,
		Item:            s.item,
	}
}

func (d *torrentDownloader) Get(item media.SearchItem, url string, destination string) (Informer, error) {
	magnet, err := d.repo.GetAvailableMagnet(item.Term)
	if err != nil {
		return nil, fmt.Errorf("could not get first available torrent: %s", err)
	}

	destFolder := path.Dir(destination)

	completion, err := torStorage.NewSqlitePieceCompletion(destFolder)
	if err != nil {
		return nil, fmt.Errorf("could not initialize sqlite completion db: %s", err)
	}

	// TODO Delete sqlite after download
	filePather := torStorage.NewFileWithCompletion(destFolder, completion)

	torrentConf := torrent.NewDefaultClientConfig()
	torrentConf.DefaultStorage = filePather
	client, err := torrent.NewClient(torrentConf)

	tor, err := client.AddMagnet(magnet)
	if err != nil {
		return nil, fmt.Errorf("could not add magnet: %s", err)
	}

	var status torrentStatus
	<-tor.GotInfo()
	for _, f := range tor.Files() {
		if f.Path() == url {
			status.file = f
			status.item = item
			status.filepath = destination

			f.Download()
			break
		}
	}

	return &status, err
}
