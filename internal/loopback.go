package internal

import (
	"github.com/akurilov/gobot/pkg/content"
	"github.com/hashicorp/golang-lru"
	"go.uber.org/zap"
	"net/url"
	"regexp"
	"strings"
)

type object struct {
}

const (
	uniqueUrlCacheSize = 0x100000 // ~ 1M
)

func initUniqueUrlCache(capacity int) *lru.Cache {
	uniqueUrls, err := lru.New(capacity)
	if err != nil {
		panic(err)
	}
	return uniqueUrls
}

var (
	linkPattern     = regexp.MustCompile("href=\"(((http[s]?://)|(www\\.))[\\S^\"]{2,256})\"")
	uniqueUrlsCache = initUniqueUrlCache(uniqueUrlCacheSize)
	placeholder     = &object{}
)

type UrlLoopBackHandler struct {
	log       *zap.Logger
	urlOutput chan<- *url.URL
}

func NewUrlLoopbackHandler(log *zap.Logger, urlOutput chan<- *url.URL) *UrlLoopBackHandler {
	return &UrlLoopBackHandler{
		log,
		urlOutput,
	}
}

func (h *UrlLoopBackHandler) Handle(c *content.Content) error {
	if !c.IsDummy() && strings.HasPrefix(c.MimeType, "text") {
		txt := string(c.Body)
		linkMatches := linkPattern.FindAllStringSubmatch(txt, 0x100)
		for _, linkMatch := range linkMatches {
			if len(linkMatch) > 1 {
				u, err := url.Parse(linkMatch[1])
				if err == nil {
					// truncate fragment and query in url
					u.Fragment = ""
					u.RawQuery = ""
					u.ForceQuery = false
					found, _ := uniqueUrlsCache.ContainsOrAdd(u, placeholder)
					if found {
						h.log.Debug("dropping non-unique url: " + u.String())
					} else {
						select {
						case h.urlOutput <- u:
						default:
							h.log.Debug("fetch queue is full, dropping the url: " + u.String())
						}
					}
				} else {
					h.log.Warn(err.Error())
				}
			}
		}
	}
	return nil
}
