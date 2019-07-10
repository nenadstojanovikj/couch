package magnet

import (
	"sort"

	"github.com/nenad/couch/pkg/storage"
)

var qualityScore = map[storage.Quality]int{
	storage.Quality4K:  4,
	storage.QualityFHD: 3,
	storage.QualityHD:  2,
	storage.QualitySD:  1,
}

var encodingScore = map[storage.Encoding]int{
	storage.Encodingx265: 4,
	storage.Encodingx264: 3,
	storage.EncodingVC1:  2,
	storage.EncodingXVID: 1,
}

func SortQuality(desc bool) ProcessFunc {
	return func(magnets []storage.Magnet) []storage.Magnet {
		sort.SliceStable(magnets, func(i, j int) bool {
			if desc {
				i, j = j, i
			}
			return qualityScore[magnets[i].Quality] < qualityScore[magnets[j].Quality]
		})
		return magnets
	}
}

func SortSize(desc bool) ProcessFunc {
	return func(magnets []storage.Magnet) []storage.Magnet {
		sort.SliceStable(magnets, func(i, j int) bool {
			if desc {
				i, j = j, i
			}
			return magnets[i].Size < magnets[j].Size
		})
		return magnets
	}
}

func SortEncoding(desc bool) ProcessFunc {
	return func(magnets []storage.Magnet) []storage.Magnet {
		sort.SliceStable(magnets, func(i, j int) bool {
			if desc {
				i, j = j, i
			}
			return encodingScore[magnets[i].Encoding] < encodingScore[magnets[j].Encoding]
		})
		return magnets
	}
}
