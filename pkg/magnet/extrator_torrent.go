package magnet

import (
	"fmt"
	"sort"

	"github.com/anacrolix/torrent"
	"github.com/nenad/couch/pkg/media"
	"github.com/nenad/couch/pkg/storage"
)

type torrentExtractor struct {
	repo *storage.MediaRepository
}

func NewTorrentExtractor(repo *storage.MediaRepository) *torrentExtractor {
	return &torrentExtractor{
		repo: repo,
	}
}

func (ex *torrentExtractor) Extract(magnet storage.Magnet) ([]string, error) {
	client, err := torrent.NewClient(nil)
	if err != nil {
		return nil, fmt.Errorf("could not create torrent client: %s", err)
	}
	defer client.Close()
	tor, err := client.AddMagnet(magnet.Location)
	if err != nil {
		return nil, fmt.Errorf("could not create torrent file: %s", err)
	}

	<-tor.GotInfo()

	var candidateFiles []*torrent.File
	for _, f := range tor.Files() {

		if !checkSuffix(f.Path()) {
			continue
		}
		candidateFiles = append(candidateFiles, f)
	}

	if len(candidateFiles) == 0 {
		return nil, fmt.Errorf("no video files found for magnet %s", magnet.Location)
	}

	// Some torrent files have
	sort.SliceStable(candidateFiles, func(i, j int) bool {
		return candidateFiles[i].Length() < candidateFiles[j].Length()
	})

	var candidates []string
	switch magnet.Item.Type {
	case media.TypeEpisode, media.TypeMovie:
		candidates = []string{candidateFiles[len(candidateFiles)-1].Path()}
	case media.TypeSeason:
		candidates = make([]string, len(candidateFiles))
		for i, r := range candidateFiles {
			candidates[i] = r.Path()
		}
	}

	return candidates, nil
}
