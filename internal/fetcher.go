package internal

import (
	"github.com/akurilov/gobot/pkg/content"
	"go.uber.org/zap"
	"net/url"
	"sync/atomic"
)

type Fetcher struct {
	log           *zap.Logger
	client        *GobotClient
	fetchCounter  *uint64
	urlInput      <-chan *url.URL
	contentOutput chan<- *content.Content
}

func NewFetcher(
	log *zap.Logger, client *GobotClient, fetchCounter *uint64, urlInput <-chan *url.URL,
	contentOutput chan<- *content.Content) *Fetcher {
	return &Fetcher{
		log,
		client,
		fetchCounter,
		urlInput,
		contentOutput,
	}
}

func (f *Fetcher) Loop() {
	for u := range f.urlInput {
		f.log.Debug("fetching: " + u.String())
		c, err := f.client.GetContent(u)
		if err == nil {
			select {
			case f.contentOutput <- c:
			default:
				f.log.Debug("parse queue is full, dropping the content")
			}
		} else {
			f.log.Debug(err.Error())
		}
		atomic.AddUint64(f.fetchCounter, 1)
	}
}
