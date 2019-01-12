package storage

import (
	"fmt"
	"github.com/nenadstojanovikj/couch/pkg/media"
	"strings"
)

func Movie(title string, year int) string {
	return fmt.Sprintf("type = '"+string(media.TypeMovie)+"' and title = '"+media.FormatMovie+"'", title, year)
}

func Episode(title string, season, episode int) string {
	return fmt.Sprintf("type = '"+string(media.TypeEpisode)+"' and title = '"+media.FormatEpisode+"'", title, season, episode)
}

func Resources(items ...media.Title) string {
	titles := make([]string, len(items))
	for i, item := range items {
		titles[i] = "'" + string(item) + "'"
	}
	return fmt.Sprintf("title in (%s)", strings.Join(titles, ","))
}
