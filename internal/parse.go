package internal

import (
	"github.com/hashicorp/golang-lru"
	"go.uber.org/zap"
	"regexp"
)

func initUniqueUrlCache(capacity int) *lru.Cache {
	uniqueUrls, err := lru.New(capacity)
	if err != nil {
		panic(err)
	}
	return uniqueUrls
}

type object struct {
}

const (
	uniqueUrlCacheSize = 0x100000 // ~ 1M
)

var (
	linkPattern     = regexp.MustCompile("href=\"(((http[s]?://)|(www\\.))[\\S^\"]{2,256})\"")
	uniqueUrlsCache = initUniqueUrlCache(uniqueUrlCacheSize)
	placeholder     = &object{}
)

func ParseLoop(log *zap.Logger, parseQueue <-chan string, fetchQueue chan<- string) {
	for txt := range parseQueue {
		linkMatches := linkPattern.FindAllStringSubmatch(txt, 0x100)
		for _, linkMatch := range linkMatches {
			if len(linkMatch) > 1 {
				url := linkMatch[1]
				if len(url) > 0 {
					url = UrlTrancate(url)
					found, _ := uniqueUrlsCache.ContainsOrAdd(url, placeholder)
					if found {
						log.Debug("dropping non-unique url: " + url)
					} else {
						select {
						case fetchQueue <- url:
						default:
							log.Debug("fetch queue is full, dropping the url: " + url)
						}
					}
				}
			}
		}
	}
}
