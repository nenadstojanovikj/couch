package media

import (
	"time"

	"github.com/nenad/trakt"
	"github.com/sirupsen/logrus"
)

func NewTraktProvider(trakt *trakt.Client) *TraktProvider {
	return &TraktProvider{trakt: trakt}
}

type TraktProvider struct {
	trakt *trakt.Client
}

func (p *TraktProvider) Poll() (metadata []SearchItem, err error) {
	var removeMeta trakt.FullMetadata

	today := time.Now().Format("2006-01-02")
	episodes, err := p.trakt.Calendar(today, 1)
	if err != nil {
		return nil, err
	}
	for _, e := range episodes {
		metadata = append(metadata, NewEpisode(e.Show.Title, e.Episode.Season, e.Episode.Number, e.Show.IDs.IMDb))
		removeMeta.Episodes = append(removeMeta.Episodes, e.Episode)
	}

	watchEpisodes, err := p.trakt.WatchlistEpisodes()
	if err != nil {
		return nil, err
	}
	for _, e := range watchEpisodes {
		metadata = append(metadata, NewEpisode(e.Show.Title, e.Episode.Season, e.Episode.Number, e.Show.IDs.IMDb))
		removeMeta.Episodes = append(removeMeta.Episodes, e.Episode)
	}

	movies, err := p.trakt.WatchlistMovies()
	if err != nil {
		return nil, err
	}
	for _, m := range movies {
		metadata = append(metadata, NewMovie(m.Movie.Title, m.Movie.Year, m.Movie.IDs.IMDb))
		removeMeta.Movies = append(removeMeta.Movies, m.Movie)
	}

	seasons, err := p.trakt.WatchlistSeasons()
	if err != nil {
		return nil, err
	}
	for _, s := range seasons {
		metadata = append(metadata, NewSeason(s.Show.Title, s.Season.Number, s.Show.IDs.IMDb))
		removeMeta.Seasons = append(removeMeta.Seasons, s.Season)
	}

	if err := p.trakt.RemoveFromWatchlist(removeMeta); err != nil {
		logrus.Errorf("could not remove metadata about items: %s", err)
		return nil, err
	}

	return metadata, nil
}

func (p *TraktProvider) Interval() time.Duration {
	return time.Minute * 15
}
