package storage

import (
	"fmt"
	"strings"
)

func Movie(title string, year int) string {
	return fmt.Sprintf("res_type = '"+string(TypeMovie)+"' and title = '"+FormatMovie+"'", title, year)
}

func TVShow(title string, season, episode int) string {
	return fmt.Sprintf("res_type = '"+string(TypeTVShow)+"' and title = '"+FormatTVShow+"'", title, season, episode)
}

func Resources(items ...MediaTitle) string {
	titles := make([]string, len(items))
	for i, item := range items {
		titles[i] = "'" + string(item) + "'"
	}
	return fmt.Sprintf("title in (%s)", strings.Join(titles, ","))
}
