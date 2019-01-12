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
	// Title represents the Movie or TVShow title. Should be generated
	// through the NewMovie or NewEpisode package methods
	Title string

	// Type is the type of media
	Type string

	Item struct {
		Title   Title
		Type    Type
		Details map[string]string
	}
)

func NewMovie(title string, year int, details map[string]string) Item {
	return Item{Title: Title(fmt.Sprintf(FormatMovie, title, year)), Type: TypeMovie, Details: details}
}

func NewEpisode(title string, season, episode int) Item {
	return Item{Title: Title(fmt.Sprintf(FormatEpisode, title, season, episode)), Type: TypeEpisode}
}
