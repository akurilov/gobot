package main

import (
	"fmt"
	"github.com/akurilov/gobot/pkg"
	"github.com/hashicorp/golang-lru"
	"go.uber.org/zap"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

const (
	contentLengthLimit = 0x100000 // 1 MB
	urlQueueSizeLimit  = 0x100000
	txtQueueSizeLimit  = 0x100000
	uniqueUrlCacheSize = 0x100000
)

type Object struct {
}

var (
	linkPattern     = regexp.MustCompile("href=\"(((http[s]?://)|(www\\.))[\\S^\"]{2,256})\"")
	client          = pkg.NewGobotClient()
	log             = initLogger()
	urlQueue        = make(chan string, urlQueueSizeLimit)
	txtQueue        = make(chan string, txtQueueSizeLimit)
	syncGroup       = sync.WaitGroup{}
	uniqueUrlsCache = initUniqueUrlCache(uniqueUrlCacheSize)
	placeholder     = &Object{}
	parallelism     = runtime.NumCPU()
)

func initLogger() *zap.Logger {
	// set up the logging
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	return logger
}

func initUniqueUrlCache(capacity int) *lru.Cache {
	uniqueUrls, err := lru.New(capacity)
	if err != nil {
		panic(err)
	}
	return uniqueUrls
}

func main() {
	defer func() {
		err := log.Sync()
		if err != nil {
			panic(err)
		}
	}()
	// handle the command line arguments
	args := os.Args
	if len(args) > 1 {
		syncGroup.Add(2)
		for i := 0; i < parallelism; i++ {
			go fetching()
			go parsing()
		}
		for _, arg := range args[1:] {
			urlQueue <- arg
		}
		syncGroup.Wait()
	} else {
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Quick and dirty internet crawler command line options\n: url1 [url2 [url3 ...]]")
}

func fetching() {
	for {
		url := <-urlQueue
		log.Info("fetching: " + url)
		txt, err := client.ContentText(url, contentLengthLimit)
		if err == nil {
			txtQueue <- txt
		} else {
			log.Warn(err.Error())
		}
	}
	syncGroup.Done()
}

func parsing() {
	for {
		txt := <-txtQueue
		linkMatches := linkPattern.FindAllStringSubmatch(txt, 0x100)
		for _, linkMatch := range linkMatches {
			if len(linkMatch) > 1 {
				url := linkMatch[1]
				if len(url) > 0 {
					url = urlTrancate(url)
					found, _ := uniqueUrlsCache.ContainsOrAdd(url, placeholder)
					if found {
						log.Debug("dropping non-unique url: " + url)
					} else {
						urlQueue <- url
					}
				}
			}
		}
	}
	syncGroup.Done()
}

func urlTrancate(url string) string {
	result := url
	anchorIdx := strings.Index(result, "#")
	if anchorIdx > 0 {
		result = result[:anchorIdx]
	}
	queryIdx := strings.Index(result, "?")
	if queryIdx > 0 {
		result = result[:queryIdx]
	}
	return result
}
