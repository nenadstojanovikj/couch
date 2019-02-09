package media

import (
	"fmt"
	"path"
	"regexp"
	"strconv"
)

const (
	// Possible types of the media item
	TypeMovie   Type = "Movie"
	TypeEpisode Type = "Episode"
	TypeSeason  Type = "Season"

	FormatMovie   string = "%s %d"
	FormatEpisode string = "%s S%02dE%02d"
	FormatSeason  string = "%s S%02d"
)

var tvShowRegex = regexp.MustCompile("(.*) S([0-9]{2})")

type (
	// Type is the type of media
	Type string

	SearchItem struct {
		Term string
		IMDb string
		Type Type
	}
)

// Path returns target download location given the base path for download, and
// the current file path (URL or location in torrent)
func (s *SearchItem) Path(basePath, filePath string) string {
	switch s.Type {
	case TypeEpisode, TypeSeason:
		matches := tvShowRegex.FindAllStringSubmatch(string(s.Term), -1)
		name := matches[0][1]
		season, _ := strconv.Atoi(matches[0][2])

		return path.Join(basePath, fmt.Sprintf("%s/Season %d/%s", name, season, path.Base(filePath)))
	default:
		return path.Join(basePath, path.Base(filePath))
	}
}

func NewMovie(title string, year int, imdb string) SearchItem {
	return SearchItem{
		Term: fmt.Sprintf(FormatMovie, title, year),
		Type: TypeMovie,
		IMDb: imdb,
	}
}

func NewEpisode(title string, season, episode int, imdb string) SearchItem {
	return SearchItem{
		Term: fmt.Sprintf(FormatEpisode, title, season, episode),
		Type: TypeEpisode,
		IMDb: imdb,
	}
}

func NewSeason(title string, season int, imdb string) SearchItem {
	return SearchItem{
		Term: fmt.Sprintf(FormatSeason, title, season),
		Type: TypeSeason,
		IMDb: imdb,
	}
}
