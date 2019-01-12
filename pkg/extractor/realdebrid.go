package extractor

import (
	"fmt"
	"github.com/nenadstojanovikj/rd"
)

type (
	RealDebridExtractor struct {
		rd *rd.RealDebrid
	}
)

func (ex *RealDebridExtractor) Extract(magnet string) (string, error) {
	// Check for maximum number of torrents

	info, err := ex.rd.Torrents.AddMagnetLinkSimple(magnet)
	if err != nil {
		return "", err
	}

	torrent, err := ex.rd.Torrents.GetTorrent(info.ID)
	if err != nil {
		return "", err
	}

	for {
		switch torrent.Status {
		case rd.StatusWaitingFiles:
			fmt.Println("HAHA")
		case rd.StatusMagnetConversion:
			fmt.Printf("MAGNET")
		}
	}
}
