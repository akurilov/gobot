package cmd

import (
	"fmt"
	"github.com/akurilov/gobot/internal"
	"go.uber.org/zap"
	"os"
	"runtime/pprof"
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

var (
	client     = internal.NewGobotClient(contentLengthLimit)
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
				internal.FetchLoop(log, client, &fetchCount, fetchQueue, parseQueue)
				syncGroup.Done()
			}()
		}
		go func() {
			internal.ParseLoop(log, parseQueue, fetchQueue)
			syncGroup.Done()
		}()
		for _, arg := range args[1:] {
			fetchQueue <- arg
		}
		f, err := os.Create("cpu.profiling")
		if err != nil {
			panic(err)
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			panic(err)
		}
		defer pprof.StopCPUProfile()
		go printingStats(f)
		syncGroup.Wait()
	} else {
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Quick and dirty internet crawler command line options\n: url1 [url2 [url3 ...]]")
}

func printingStats(w *os.File) {
	sugar := log.Sugar()
	for {
		w.Sync()
		time.Sleep(statsOutputPeriod * time.Second)
		c := atomic.LoadUint64(&fetchCount)
		d := time.Since(startTime)
		sugar.Infof(
			"URL queue length: %d, parse queue length: %d, fetch count: %d, mean rate: %f", len(fetchQueue),
			len(parseQueue), c, float64(c)/d.Seconds())
	}
}
