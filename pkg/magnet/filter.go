package magnet

import (
	"github.com/nenad/couch/pkg/storage"
)

type (
	ProcessFunc func([]storage.Magnet) []storage.Magnet
)

func FilterQuality(min, max storage.Quality) ProcessFunc {
	return func(magnets []storage.Magnet) (new []storage.Magnet) {
		for _, m := range magnets {
			score := qualityScore[m.Quality]
			minScore := qualityScore[min]
			maxScore := qualityScore[max]
			if score >= minScore && score <= maxScore {
				new = append(new, m)
			}
		}

		return new
	}
}

func FilterEncoding(encodings ...storage.Encoding) ProcessFunc {
	allowedEncodings := make(map[storage.Encoding]bool)
	for _, e := range encodings {
		allowedEncodings[e] = true
	}

	return func(magnets []storage.Magnet) (new []storage.Magnet) {
		for _, m := range magnets {
			if _, ok := allowedEncodings[m.Encoding]; ok {
				new = append(new, m)
			}
		}

		return new
	}
}
