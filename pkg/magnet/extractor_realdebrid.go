package magnet

import (
	"fmt"
	"github.com/nenadstojanovikj/rd"
	"sort"
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

func (ex *RealDebridExtractor) Extract(magnet string) (string, error) {
	// Check for maximum number of torrents

	info, err := ex.debrid.Torrents.AddMagnetLinkSimple(magnet)
	if err != nil {
		return "", err
	}

	var torrent rd.TorrentInfo

loop:
	for {
		torrent, err = ex.debrid.Torrents.GetTorrent(info.ID)
		if err != nil {
			return "", fmt.Errorf("extractor: could not get info about torrent %s: %s", info.ID, err)
		}

		switch torrent.Status {
		// Select Video files and start download
		case rd.StatusWaitingFiles:
			// TODO Check if we need to download the whole torrent
			fileIDs := extractFileIDs(torrent, ex.downloadAll)
			if err := ex.debrid.Torrents.SelectFilesFromTorrent(torrent.ID, fileIDs); err != nil {
				return "", fmt.Errorf("extractor: could not select files for download: %s", err)
			}

		// Return error if we cannot download the torrent for some reason
		case rd.StatusDead:
		case rd.StatusMagnetError:
		case rd.StatusVirus:
		case rd.StatusError:
			return "", fmt.Errorf("extractor: could not download torrent %s, status %s", torrent.ID, torrent.Status)

		// Only exit the loop when the torrent is successfully downloaded
		case rd.StatusDownloaded:
			break loop
		}
		time.Sleep(ex.pollInterval)
	}

	torrentLink := torrent.Links[0]
	linkInfo, err := ex.debrid.Unrestrict.SimpleUnrestrict(torrentLink)
	if err != nil {
		return "", fmt.Errorf("extractor: could not unrestrict link %s: %s", linkInfo.Download, err)
	}

	return linkInfo.Download, nil
}

func extractFileIDs(torrentInfo rd.TorrentInfo, extractAll bool) []int {
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

	sort.SliceStable(candidateFiles, func(i, j int) bool {
		return candidateFiles[i].Bytes > candidateFiles[j].Bytes
	})

	return []int{candidateFiles[len(candidateFiles)-1].ID}
}

func checkSuffix(path string) bool {
	if len(path) < 5 {
		return false
	}
	_, ok := allowedExtensions[path[len(path)-4:]]
	return ok
}
