package magnet

import (
	"github.com/nenadstojanovikj/couch/pkg/storage"
	"sort"
)

type (
	Sort func([]storage.Magnet)
)

var qualityScore = map[storage.Quality]int{
	storage.Quality4K:  4,
	storage.QualityFHD: 3,
	storage.QualityHD:  2,
	storage.QualitySD:  1,
}

var encodingScore = map[storage.Encoding]int{
	storage.EncodingHEVC: 5,
	storage.Encodingx265: 4,
	storage.Encodingx264: 3,
	storage.EncodingVC1:  2,
	storage.EncodingXVID: 1,
}

func SortQuality(desc bool) Sort {
	return func(magnets []storage.Magnet) {
		sort.Slice(magnets, func(i, j int) bool {
			if desc {
				i, j = j, i
			}
			return qualityScore[magnets[i].Quality] < qualityScore[magnets[j].Quality]
		})
	}
}

func SortEncoding(desc bool) Sort {
	return func(magnets []storage.Magnet) {
		sort.Slice(magnets, func(i, j int) bool {
			if desc {
				i, j = j, i
			}
			return encodingScore[magnets[i].Encoding] < encodingScore[magnets[j].Encoding]
		})
	}
}
