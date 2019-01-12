package magnet

import (
	"fmt"
	"github.com/nenadstojanovikj/couch/pkg/media"
	"github.com/nenadstojanovikj/couch/pkg/storage"
	"github.com/qopher/go-torrentapi"
)

type (
	RarbgScraper struct {
		client *torrentapi.API
	}
)

func NewRarbgScraper() (*RarbgScraper, error) {
	api, err := torrentapi.New("couch")
	if err != nil {
		return nil, err
	}
	return &RarbgScraper{client: api}, nil
}

func (s *RarbgScraper) Scrape(item media.Item) ([]storage.Magnet, error) {
	query := s.client.SearchString(string(item.Title))
	query.Format("json_extended")
	switch item.Type {
	case media.TypeEpisode:
		query.
			Category(18). // TV Episodes
			Category(41). // TV HD Episodes
			Category(49)  // TV UHD Episodes
	case media.TypeMovie:
		query.
			Category(14). // Movies/XVID
			Category(17). // Movies/x264
			Category(44). // Movies/x264/1080
			Category(45). // Movies/x264/720
			Category(47). // Movies/x264/3D
			Category(50). // Movies/x264/4k
			Category(51). // Movies/x265/4k
			Category(52). // Movies/x265/4k/HDR
			Category(42). // Movies/Full BD
			Category(46)  // Movies/BD Remux
	}

	results, err := query.Search()
	if err != nil {
		return nil, err
	}
	for _, r := range results {
		fmt.Printf("%+v\n", r)
	}

	return nil, nil
}
