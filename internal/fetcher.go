package internal

import (
	"github.com/akurilov/gobot/pkg/content"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
	"net/url"
)

var (
	fetchCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "gobot_fetch_count_total",
			Help: "The total number of fetched contents ",
		})
)

type Fetcher struct {
	log           *zap.Logger
	client        *GobotClient
	urlInput      <-chan *url.URL
	contentOutput chan<- *content.Content
}

func NewFetcher(
	log *zap.Logger, client *GobotClient, urlInput <-chan *url.URL, contentOutput chan<- *content.Content) *Fetcher {
	return &Fetcher{
		log,
		client,
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
		fetchCounter.Inc()
	}
}
