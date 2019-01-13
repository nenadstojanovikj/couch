package mediaprovider

import (
	"github.com/nenadstojanovikj/couch/pkg/media"
	"time"
)

type Poller interface {
	Poll() ([]media.Item, error)
	Interval() time.Duration
}
