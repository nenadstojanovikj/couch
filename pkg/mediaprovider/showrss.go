package mediaprovider

import (
	"fmt"
	"github.com/nenadstojanovikj/couch/pkg/media"
	"github.com/nenadstojanovikj/showrss-go"
	"sort"
	"time"
)

type ShowRSSProvider struct {
	client      *showrss.Client
	personalUrl string
	interval    time.Duration
}

func NewShowRSSProvider(interval time.Duration, url string, rss *showrss.Client) *ShowRSSProvider {
	return &ShowRSSProvider{
		client:      rss,
		personalUrl: url,
		interval:    interval,
	}
}

func (p *ShowRSSProvider) Poll() (titles []media.Item, err error) {
	items, err := p.client.GetPersonalEpisodes(p.personalUrl)
	if err != nil {
		return
	}
	items = filterMaxDefinitionEpisodes(items)

	titles = make([]media.Item, len(items))
	for i, episode := range items {
		titles[i] = media.NewEpisode(episode.ShowName, episode.Season, episode.Episode)
	}

	return titles, nil
}

func (p *ShowRSSProvider) Interval() time.Duration {
	return p.interval
}

func filterMaxDefinitionEpisodes(episodes []showrss.Episode) (filtered []showrss.Episode) {
	aggregated := make(map[string]showrss.Episode)
	for _, e := range episodes {
		key := fmt.Sprintf("%s%dx%d", e.ShowName, e.Season, e.Episode)

		// If we have no items, add it and continue loop
		existingEpisode, ok := aggregated[key]
		if !ok {
			aggregated[key] = e
			continue
		}

		if isQualityBetter(existingEpisode, e) {
			aggregated[key] = e
		}
	}

	for _, e := range aggregated {
		filtered = append(filtered, e)
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].ShowName != filtered[j].ShowName {
			return filtered[i].ShowName < filtered[j].ShowName
		}

		if filtered[i].Season != filtered[j].Season {
			return filtered[i].Season < filtered[j].Season
		}

		return filtered[i].Episode < filtered[j].Episode
	})

	return filtered
}

func isQualityBetter(old showrss.Episode, new showrss.Episode) bool {
	q := map[showrss.Quality]int{
		showrss.QualityFullHD: 2,
		showrss.QualityHD:     1,
		showrss.QualitySD:     0,
	}

	return q[new.Quality] > q[old.Quality]
}
