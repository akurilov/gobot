package pkg

import (
	"go.uber.org/zap"
	"sync/atomic"
)

func FetchLoop(
	log *zap.Logger, client *GobotClient, fetchCounter *uint64, fetchQueue <-chan string, parseQueue chan<- string) {
	for url := range fetchQueue {
		log.Debug("fetching: " + url)
		txt, err := client.ContentText(url)
		if err == nil {
			select {
			case parseQueue <- txt:
			default:
				log.Debug("parse queue is full, dropping the content")
			}
		} else {
			log.Debug(err.Error())
		}
		atomic.AddUint64(fetchCounter, 1)
	}
}
