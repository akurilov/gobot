package main

import (
	"fmt"
	"github.com/akurilov/gobot/pkg"
	"go.uber.org/zap"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const (
	contentLengthLimit  = 0x100000 // 1 MB
	fetchQueueSizeLimit = 0x10000
	parseQueueSizeLimit = 0x10000
	fetchConcurrency    = 100
	statsOutputPeriod   = 10
)

type Object struct {
}

var (
	client     = pkg.NewGobotClient(contentLengthLimit)
	log        = initLogger()
	fetchQueue = make(chan string, fetchQueueSizeLimit)
	parseQueue = make(chan string, parseQueueSizeLimit)
	syncGroup  = sync.WaitGroup{}
	fetchCount = uint64(0)
	startTime  = time.Now()
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
	defer close(fetchQueue)
	defer close(parseQueue)
	args := os.Args
	if len(args) > 1 {
		syncGroup.Add(1)
		for i := 0; i < fetchConcurrency; i++ {
			go func() {
				pkg.FetchLoop(log, client, &fetchCount, fetchQueue, parseQueue)
				syncGroup.Done()
			}()
		}
		go func() {
			pkg.ParseLoop(log, parseQueue, fetchQueue)
			syncGroup.Done()
		}()
		go printingStats()
		for _, arg := range args[1:] {
			fetchQueue <- arg
		}
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
			"URL queue length: %d, parse queue length: %d, fetch count: %d, mean rate: %f", len(fetchQueue),
			len(parseQueue), c, float64(c)/d.Seconds())
	}
}
