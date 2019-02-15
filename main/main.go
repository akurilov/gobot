package main

import (
	"fmt"
	"github.com/akurilov/gobot/pkg"
	"go.uber.org/zap"
	"os"
)

const (
	contentLengthLimit = 0x100000 // 1 MB
)

func main() {

	// set up the logging
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() {
		err := logger.Sync()
		if err != nil {
			panic(err)
		}
	}()

	// set up the client
	client := pkg.NewGobotClient()

	// handle the command line arguments
	args := os.Args
	if len(args) > 1 {
		handleUrls(logger, client, "", args[1:])
	} else {
		printUsage()
	}
}

func printUsage() {
	fmt.Printf("Quick and dirty internet crawler command line options\n: url1 [url2 [url3 ...]]")
}

func handleUrls(logger *zap.Logger, client *pkg.GobotClient, parentUrl string, urls []string) {
	exitChan := make(chan error, len(urls))
	submitCount := 0
	for _, url := range urls {
		go client.ContentText(url, contentLengthLimit, handleContent, exitChan)
		submitCount++
	}
	for i := 0; i < submitCount; i++ {
		err := <-exitChan
		if err != nil {
			logger.Error(err.Error())
		}
	}

}

func handleContent(url string, txt string) error {

}
