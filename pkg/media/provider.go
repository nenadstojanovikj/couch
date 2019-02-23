package media

import (
	"time"
)

type Provider interface {
	Poll() ([]SearchItem, error)
	Interval() time.Duration
}
