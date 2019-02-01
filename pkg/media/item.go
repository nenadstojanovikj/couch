package media

import (
	"fmt"
)

const (
	// Possible types of the media item
	TypeMovie   Type = "Movie"
	TypeEpisode Type = "Episode"

	FormatMovie   string = "%s %d"
	FormatEpisode string = "%s S%02dE%02d"
)

type (
	// Type is the type of media
	Type string

	SearchItem struct {
		UniqueTitle string
		IMDb        string
		Type        Type
	}
)

func NewMovie(title string, year int, imdb string) SearchItem {
	return SearchItem{
		UniqueTitle: fmt.Sprintf(FormatMovie, title, year),
		Type:        TypeMovie,
		IMDb:        imdb,
	}
}

func NewEpisode(title string, season, episode int, imdb string) SearchItem {
	return SearchItem{
		UniqueTitle: fmt.Sprintf(FormatEpisode, title, season, episode),
		Type:        TypeEpisode,
		IMDb:        imdb,
	}
}
