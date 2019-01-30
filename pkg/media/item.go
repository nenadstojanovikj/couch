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
	// UniqueTitle represents the Movie or TVShow title. Should be generated
	// through the NewMovie or NewEpisode package methods
	Title string

	// Type is the type of media
	Type string

	Metadata struct {
		Title       string
		UniqueTitle string
		Type        Type

		// Used if data is about Movie
		Year int

		// Used if data is about TV Show
		Season  int
		Episode int
	}
)

func NewMovie(title string, year int) Metadata {
	return Metadata{
		Title:       title,
		UniqueTitle: fmt.Sprintf(FormatMovie, title, year),
		Type:        TypeMovie,
		Year:        year,
	}
}

func NewEpisode(title string, season, episode int) Metadata {
	return Metadata{
		Title:       title,
		UniqueTitle: fmt.Sprintf(FormatEpisode, title, season, episode),
		Type:        TypeEpisode,
		Episode:     episode,
		Season:      season,
	}
}
