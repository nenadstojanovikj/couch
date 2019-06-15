package notifications

import (
	"github.com/nenad/couch/pkg/media"
)

type NoopNotifier struct {
}

func (n *NoopNotifier) OnQueued(item media.SearchItem) error {
	return nil
}

func (n *NoopNotifier) OnFinish(item media.SearchItem) error {
	return nil
}
