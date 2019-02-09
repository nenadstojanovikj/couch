package magnet

import (
	"fmt"
	"github.com/nenadstojanovikj/couch/pkg/media"
	"github.com/nenadstojanovikj/couch/pkg/storage"
	"github.com/nenadstojanovikj/rd"
	"github.com/sirupsen/logrus"
	"sort"
	"strings"
	"time"
)

type (
	RealDebridExtractor struct {
		debrid       *rd.RealDebrid
		pollInterval time.Duration
		downloadAll  bool
	}
)

func NewRealDebridExtractor(debrid *rd.RealDebrid, pollInterval time.Duration, downloadAll bool) *RealDebridExtractor {
	return &RealDebridExtractor{
		debrid:       debrid,
		pollInterval: pollInterval,
		downloadAll:  downloadAll,
	}
}

func (ex *RealDebridExtractor) Extract(magnet storage.Magnet) ([]string, error) {
	// Check for maximum number of torrents

	info, err := ex.AddOrGetTorrentUrl(magnet.Location)
	if err != nil {
		return nil, fmt.Errorf("extractor: could not add torrent: %s", err)
	}

	var torrent rd.TorrentInfo

loop:
	for {
		torrent, err = ex.debrid.Torrents.GetTorrent(info.ID)
		if err != nil {
			return nil, fmt.Errorf("extractor: could not get info about torrent %s: %s", info.ID, err)
		}

		switch torrent.Status {
		// Select Video files and start download
		case rd.StatusWaitingFiles:
			// TODO Check if we need to download the whole torrent
			fileIDs := ex.extractFileIDs(torrent, magnet, ex.downloadAll)
			if err := ex.debrid.Torrents.SelectFilesFromTorrent(torrent.ID, fileIDs); err != nil {
				return nil, fmt.Errorf("extractor: could not select files for download: %s", err)
			}

		// Return error if we cannot download the torrent for some reason
		case rd.StatusDead, rd.StatusMagnetError, rd.StatusVirus, rd.StatusError:
			return nil, fmt.Errorf("extractor: could not download torrent %s, status %s", torrent.ID, torrent.Status)

		// Only exit the loop when the torrent is successfully downloaded
		case rd.StatusDownloaded:
			if err := ex.debrid.Torrents.Delete(torrent.ID); err != nil {
				logrus.Warnf("failed to delete torrent in realdebrid: %s", err)
			}
			break loop
		}
		logrus.Debugf("waiting on RealDebrid, progress %d (%s)", torrent.Progress, magnet.Item.Term)
		time.Sleep(ex.pollInterval)
	}

	links := make([]string, len(torrent.Links))
	for i, link := range torrent.Links {
		linkInfo, err := ex.debrid.Unrestrict.SimpleUnrestrict(link)
		if err != nil {
			return nil, fmt.Errorf("extractor: could not unrestrict link %s: %s", linkInfo.Download, err)
		}

		links[i] = linkInfo.Download
	}

	return links, nil
}

func (ex *RealDebridExtractor) extractFileIDs(torrentInfo rd.TorrentInfo, magnet storage.Magnet, extractAll bool) []int {
	// Get everything
	if extractAll {
		ids := make([]int, len(torrentInfo.Files))
		for i, r := range torrentInfo.Files {
			ids[i] = r.ID
		}
		return ids
	}

	// Get only the biggest video file
	var candidateFiles []rd.File
	for _, f := range torrentInfo.Files {
		if !checkSuffix(f.Path) {
			continue
		}
		candidateFiles = append(candidateFiles, f)
	}

	// Some torrent files have
	sort.SliceStable(candidateFiles, func(i, j int) bool {
		return candidateFiles[i].Bytes < candidateFiles[j].Bytes
	})

	switch magnet.Item.Type {
	case media.TypeEpisode, media.TypeMovie:
		return []int{candidateFiles[len(candidateFiles)-1].ID}
	case media.TypeSeason:
		ids := make([]int, len(candidateFiles))
		for i, r := range candidateFiles {
			ids[i] = r.ID
		}
		return ids
	}

	return []int{0}
}

func (ex *RealDebridExtractor) AddOrGetTorrentUrl(magnet string) (info rd.TorrentUrlInfo, err error) {
	info, err = ex.debrid.Torrents.AddMagnetLinkSimple(magnet)
	if err == nil {
		return info, nil
	}
	logrus.Warnf("could not add magnet, probably one is already active: %s", err)

	torrents, err := ex.debrid.Torrents.GetTorrents()
	if err != nil {
		return info, fmt.Errorf("could not get list of active torrents: %s", err)
	}

	for _, t := range torrents {
		if strings.Contains(magnet, t.Hash) {
			return rd.TorrentUrlInfo{
				ID:  t.ID,
				URI: "https://api.real-debrid.com/rest/1.0/torrents/info/" + t.ID,
			}, nil
		}
	}

	return info, fmt.Errorf("could not add magnet, and magnet is not active")
}

func checkSuffix(path string) bool {
	if len(path) < 5 {
		return false
	}
	_, ok := allowedExtensions[path[len(path)-4:]]
	return ok
}
