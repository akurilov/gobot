package main

import (
	"fmt"
	"github.com/akurilov/gobot/internal"
	"github.com/akurilov/gobot/pkg/content"
	"github.com/akurilov/gobot/pkg/content/handle"
	"go.uber.org/zap"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const (
	contentLengthLimit    = 0x100000 // 1 MB
	urlQueueSizeLimit     = 0x10000
	contentQueueSizeLimit = 0x1000
	fetchConcurrency      = 0x100 // up to 256 simultaneous GET requests
	statsOutputPeriod     = 10
)

var (
	client       = internal.NewGobotClient(contentLengthLimit)
	log          = initLogger()
	urlQueue     = make(chan *url.URL, urlQueueSizeLimit)
	contentQueue = make(chan *content.Content, contentQueueSizeLimit)
	syncGroup    = sync.WaitGroup{}
	fetchCount   = uint64(0)
	startTime    = time.Now()
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
		for i := 0; i < fetchConcurrency; i++ {
			fetcher := internal.NewFetcher(log, client, &fetchCount, urlQueue, contentQueue)
			syncGroup.Add(1)
			go func() {
				fetcher.Loop()
				syncGroup.Done()
			}()
		}
		delegateHandlers := make([]handle.ContentHandle, 0, 1)
		loopBackHandler := internal.NewUrlLoopbackHandler(log, urlQueue)
		delegateHandlers = append(delegateHandlers, loopBackHandler)
		handler := internal.NewHandler(log, contentQueue, delegateHandlers)
		go func() {
			handler.Loop()
			syncGroup.Done()
		}()
		for _, arg := range args[1:] {
			u, err := url.Parse(arg)
			if err != nil {
				panic(err)
			}
			urlQueue <- u
		}
		go printingStats()
		syncGroup.Wait()
	} else {
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Quick and dirty internet crawler command line options\n: url1 [url2 [url3 ...]]")
}

func printingStats() {
	sugar := log.Sugar()
	for {
		time.Sleep(statsOutputPeriod * time.Second)
		c := atomic.LoadUint64(&fetchCount)
		d := time.Since(startTime)
		sugar.Infof(
			"URL queue length: %d, parse queue length: %d, fetch count: %d, mean rate: %f", len(urlQueue),
			len(contentQueue), c, float64(c)/d.Seconds())
	}
}
