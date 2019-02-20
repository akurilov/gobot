package cmd

import (
	"fmt"
	"github.com/akurilov/gobot/internal"
	"github.com/akurilov/gobot/pkg/content"
	"github.com/akurilov/gobot/pkg/content/handle"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"os"
)

const (
	contentLengthLimit    = 0x100000 // 1 MB
	urlQueueSizeLimit     = 0x10000
	contentQueueSizeLimit = 0x1000
	fetchConcurrency      = 0x100 // up to 256 simultaneous GET requests
)

var (
	client       = internal.NewGobotClient(contentLengthLimit)
	log          = initLogger()
	urlQueue     = make(chan *url.URL, urlQueueSizeLimit)
	contentQueue = make(chan *content.Content, contentQueueSizeLimit)
)

func initLogger() *zap.Logger {
	// set up the logging
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	return logger
}

func main() {
	defer func() {
		err := log.Sync()
		if err != nil {
			panic(err)
		}
	}()
	defer close(urlQueue)
	defer close(contentQueue)
	args := os.Args
	if len(args) > 1 {
		run(args)
	} else {
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Quick and dirty internet crawler command line options\n: url1 [url2 [url3 ...]]")
}

func run(args []string) {
	// start the content fetching
	for i := 0; i < fetchConcurrency; i++ {
		fetcher := internal.NewFetcher(log, client, urlQueue, contentQueue)
		go fetcher.Loop()
	}
	// start the content handling
	delegateHandlers := make([]handle.ContentHandle, 0, 1)
	loopBackHandler := internal.NewUrlLoopbackHandler(log, urlQueue)
	delegateHandlers = append(delegateHandlers, loopBackHandler)
	handler := internal.NewHandler(log, contentQueue, delegateHandlers)
	go handler.Loop()
	// handle command line args
	for _, arg := range args[1:] {
		u, err := url.Parse(arg)
		if err != nil {
			panic(err)
		}
		urlQueue <- u
	}
	// expose the metrics for prometheus
	urlQueueSizeGaugeOpts := prometheus.GaugeOpts{
		Name: "gobot_url_queue_size",
		Help: "URLs queue size",
	}
	prometheus.MustRegister(
		prometheus.NewGaugeFunc(
			urlQueueSizeGaugeOpts,
			func() float64 {
				return float64(len(urlQueue))
			}))

	http.Handle("/metrics", promhttp.Handler())
	log.Error(http.ListenAndServe(":2112", nil).Error())
}
