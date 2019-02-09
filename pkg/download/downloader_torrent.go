package download

import (
	"fmt"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	storage2 "github.com/anacrolix/torrent/storage"
	"github.com/nenadstojanovikj/couch/pkg/config"
	"github.com/nenadstojanovikj/couch/pkg/media"
	"github.com/nenadstojanovikj/couch/pkg/storage"
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

	usefulBytes := s.file.Torrent().Stats().BytesReadUsefulData

	if s.file.Length() == (&usefulBytes).Int64() {
		done = true
	}

	return &Info{
		Url:             s.file.Path(),
		Error:           err,
		IsDone:          done,
		TotalBytes:      s.file.Length(),
		DownloadedBytes: (&usefulBytes).Int64(),
		Filepath:        s.filepath,
		Item:            s.item,
	}
}

func (d *torrentDownloader) Get(item media.SearchItem, url string, destination string) (Informer, error) {
	// TODO New torrent client for each new torrent
	magnet, err := d.repo.GetAvailableMagnet(item.Term)
	if err != nil {
		return nil, fmt.Errorf("could not get first available torrent: %s", err)
	}

	filePather := storage2.NewFileWithCustomPathMaker("/", func(baseDir string, info *metainfo.Info, infoHash metainfo.Hash) string {
		switch item.Type {
		case media.TypeEpisode, media.TypeSeason:
			return path.Dir(item.Path(d.config.TVShowsPath, url))
		default:
			return path.Dir(item.Path(d.config.MoviesPath, url))
		}
	})

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
