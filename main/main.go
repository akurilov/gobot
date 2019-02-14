package main

import (
	"fmt"
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
		err := handle("", args[1:])
		if err != nil {
			panic(err)
		}
	} else {
		printUsage()
	}
}

func printUsage() {
	fmt.Printf("Quick and dirty internet crawler command line options\n: url1 [url2 [url3 ...]]")
}

func handle(parentUrl string, urls []string) error {
	for _, url := range urls {

		// TODO
	}
	return nil
}

func content(url string) ([]byte, error) {

}
