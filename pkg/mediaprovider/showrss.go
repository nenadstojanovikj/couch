package mediaprovider

import (
	"github.com/nenadstojanovikj/couch/pkg/media"
	"github.com/nenadstojanovikj/couch/pkg/rss"
	"github.com/nenadstojanovikj/showrss-go"
	"time"
)

type ShowRSSProvider struct {
	client      *showrss.Client
	personalUrl string
	timeout     time.Duration
}

func NewShowRSSProvider(timeout time.Duration, url string, rss *showrss.Client) *ShowRSSProvider {
	return &ShowRSSProvider{
		client:      rss,
		personalUrl: url,
		timeout:     timeout,
	}
}

func (p *ShowRSSProvider) Poll() (titles []media.Item, err error) {
	items, err := p.client.GetPersonalEpisodes(p.personalUrl)
	if err != nil {
		return
	}
	items = rss.FilterMaxDefinitionEpisodes(items)

	titles = make([]media.Item, len(items))
	for i, episode := range items {
		titles[i] = media.NewEpisode(episode.ShowName, episode.Season, episode.Episode)
	}

	return titles, nil
}

func (p *ShowRSSProvider) Timeout() time.Duration {
	return p.timeout
}
