package magnet

import (
	"github.com/nenadstojanovikj/couch/pkg/media"
	"github.com/nenadstojanovikj/couch/pkg/storage"
	"github.com/qopher/go-torrentapi"
	"github.com/sirupsen/logrus"
	"regexp"
)

type (
	RarbgScraper struct {
		client *torrentapi.API
	}
)

func NewRarbgScraper() (*RarbgScraper, error) {
	api, err := torrentapi.New("couch")
	if err != nil {
		return nil, err
	}
	return &RarbgScraper{client: api}, nil
}

func (s *RarbgScraper) Scrape(item media.Item) ([]storage.Magnet, error) {
	query := s.client.SearchString(string(item.Title))
	query.Format("json_extended")
	switch item.Type {
	case media.TypeEpisode:
		query.
			Category(18). // TV Episodes
			Category(41). // TV HD Episodes
			Category(49)  // TV UHD Episodes
	case media.TypeMovie:
		query.
			Category(14). // Movies/XVID
			Category(17). // Movies/x264
			Category(44). // Movies/x264/1080
			Category(45). // Movies/x264/720
			Category(47). // Movies/x264/3D
			Category(50). // Movies/x264/4k
			Category(51). // Movies/x265/4k
			Category(52). // Movies/x265/4k/HDR
			Category(42). // Movies/Full BD
			Category(46)  // Movies/BD Remux
	}

	results, err := query.Search()
	if err != nil {
		return nil, err
	}
	for _, r := range results {
		logrus.Debugf("found magnet %s for %q", r.Download, item.Title)
	}

	magnets := make([]storage.Magnet, len(results))
	for i, m := range results {
		magnets[i].Location = m.Download
		magnets[i].Quality = parseQuality(m)
		magnets[i].Item = item
		magnets[i].Encoding = parseEncoding(m)
		magnets[i].Size = m.Size
	}

	return magnets, nil
}

var categoryQuality = map[string]storage.Quality{
	"Movies/XVID":        storage.QualitySD,
	"Movies/x264":        storage.QualitySD,
	"Movies/x264/720":    storage.QualityHD,
	"Movies/x264/1080":   storage.QualityFHD,
	"Movies/x264/4k":     storage.Quality4K,
	"Movies/x265/4k":     storage.Quality4K,
	"Movies/x265/4k/HDR": storage.Quality4K,
}

var categoryEncoding = map[string]storage.Encoding{
	"Movies/XVID":        storage.EncodingXVID,
	"Movies/x264":        storage.Encodingx264,
	"Movies/x264/720":    storage.Encodingx264,
	"Movies/x264/1080":   storage.Encodingx264,
	"Movies/x264/4k":     storage.Encodingx264,
	"Movies/x265/4k":     storage.Encodingx265,
	"Movies/x265/4k/HDR": storage.Encodingx265,
}

var qualityRegex = regexp.MustCompile("2160p|1080p|720p")
var encodingRegex = regexp.MustCompile("[xXhH]264|[xXhH]265|HEVC|[xX][vV][iI][dD]")

func parseQuality(result torrentapi.TorrentResult) storage.Quality {
	if q, ok := categoryQuality[result.Category]; ok {
		return q
	}

	matches := qualityRegex.FindAllStringSubmatch(result.Title, -1)
	if len(matches) != 1 {
		return storage.QualitySD
	}

	return storage.Quality(matches[0][0])
}

func parseEncoding(result torrentapi.TorrentResult) storage.Encoding {
	if q, ok := categoryEncoding[result.Category]; ok {
		return q
	}

	matches := encodingRegex.FindAllStringSubmatch(result.Title, -1)
	if len(matches) != 1 {
		return storage.Encodingx264
	}

	return storage.Encoding(matches[0][0])
}
