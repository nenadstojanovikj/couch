package magnet

import (
	"fmt"
	"regexp"

	"github.com/nenad/couch/pkg/media"
	"github.com/nenad/couch/pkg/storage"
	torrentapi "github.com/qopher/go-torrentapi"
	"github.com/sirupsen/logrus"
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

func (s *RarbgScraper) Scrape(item media.SearchItem) ([]storage.Magnet, error) {
	query := s.client

	// RARBG has some weird algorithm for searching movies by titles, and
	// usually only an IMDb identifier yields better results
	if item.IMDb != "" && item.Type == media.TypeMovie {
		query = s.client.SearchIMDb(item.IMDb)
	} else {
		query = s.client.SearchString(fmt.Sprintf("%q", item.Term))

		if item.IMDb != "" {
			query = s.client.SearchIMDb(item.IMDb)
		}
	}

	query.Format("json_extended")
	switch item.Type {
	case media.TypeEpisode, media.TypeSeason:
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

	magnets := s.filterByType(item, results)

	return magnets, nil
}

func (s *RarbgScraper) filterByType(item media.SearchItem, results torrentapi.TorrentResults) []storage.Magnet {
	var filteredResults torrentapi.TorrentResults

	if item.Type == media.TypeSeason {
		for _, r := range results {
			if r.EpisodeInfo.EpisodeNum != "1000000" {
				continue
			}
			filteredResults = append(filteredResults, r)
		}
		results = filteredResults
	}

	magnets := make([]storage.Magnet, len(results))
	for i, m := range results {
		magnets[i].Location = m.Download
		magnets[i].Quality = parseQuality(m)
		magnets[i].Item = item
		magnets[i].Encoding = parseEncoding(m)
		magnets[i].Size = m.Size
		magnets[i].Seeders = m.Seeders

		logrus.Debugf("found magnet %s for %q", m.Download, item.Term)
	}

	return magnets
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
var encodingRegexes = map[storage.Encoding]*regexp.Regexp{
	storage.Encodingx264: regexp.MustCompile("[xXhH]264"),
	storage.Encodingx265: regexp.MustCompile("[xXhH]265|hevc|HEVC"),
	storage.EncodingXVID: regexp.MustCompile("[xX][vV][iI][dD]"),
	storage.EncodingVC1:  regexp.MustCompile("vc1|VC1|VC-1|vc-1"),
}

func parseQuality(result torrentapi.TorrentResult) storage.Quality {
	if q, ok := categoryQuality[result.Category]; ok {
		return q
	}

	matches := qualityRegex.FindAllStringSubmatch(result.Title, -1)
	if len(matches) != 1 {
		return storage.QualitySD
	}

	qualityStr := matches[0][0]
	switch qualityStr {
	case "720p":
		return storage.QualityHD
	case "1080p":
		return storage.QualityFHD
	case "2160p":
		return storage.Quality4K
	default:
		return storage.Quality(qualityStr)
	}
}

func parseEncoding(result torrentapi.TorrentResult) storage.Encoding {
	if q, ok := categoryEncoding[result.Category]; ok {
		return q
	}

	for enc, regex := range encodingRegexes {
		matches := regex.FindAllStringSubmatch(result.Title, -1)
		if len(matches) >= 1 {
			return enc
		}
	}

	return storage.Encodingx264
}
