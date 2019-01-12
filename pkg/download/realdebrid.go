package download

import (
	"fmt"
	"github.com/nenadstojanovikj/couch/pkg/media"
	"github.com/nenadstojanovikj/rd"
	"sort"
)

type (
	MagnetTitle struct {
		Title  media.Title
		Magnet string
	}

	TorrentIDTitle struct {
		Title     media.Title
		TorrentID string
	}
)

var allowedExtensions = map[string]bool{".mp4": true, ".mkv": true, ".avi": true, ".m4v": true, ".wmv": true}

func checkSuffix(path string) bool {
	if len(path) < 5 {
		return false
	}
	_, ok := allowedExtensions[path[len(path)-4:]]
	return ok
}

func startDownload(rdTorrent rd.TorrentService, item MagnetTitle) (t TorrentIDTitle, err error) {
	urlInfo, err := rdTorrent.AddMagnetLinkSimple(item.Magnet)
	if err != nil {
		fmt.Printf("Failed to add magnet %s: %s\n", item.Magnet, err)
		return
	}
	torrentInfo, err := rdTorrent.GetTorrent(urlInfo.ID)
	if err != nil {
		fmt.Printf("Failed to get info about torrent %s: %s\n", urlInfo.ID, err)
		return
	}

	if torrentInfo.Status != rd.StatusWaitingFiles {
		fmt.Printf("Skipping torrent in status: %s\n", torrentInfo.Status)
		return
	}

	fmt.Printf("Starting torrent: %s - Status: %s ", torrentInfo.ID, torrentInfo.Status)
	var candidateFiles []rd.File
	for _, f := range torrentInfo.Files {
		if !checkSuffix(f.Path) {
			continue
		}
		candidateFiles = append(candidateFiles, f)
	}

	if len(candidateFiles) == 0 {
		fmt.Println("No files, skipping")
		return
	}

	sort.SliceStable(candidateFiles, func(i, j int) bool {
		return candidateFiles[i].Bytes < candidateFiles[j].Bytes
	})

	file := candidateFiles[len(candidateFiles)-1]

	err = rdTorrent.SelectFilesFromTorrent(torrentInfo.ID, []int{file.ID})
	if err != nil {
		fmt.Printf("Error while selecting files: %s\n", err)
		return
	}

	return TorrentIDTitle{Title: item.Title, TorrentID: urlInfo.ID}, err
}
