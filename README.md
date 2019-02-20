[![Go Report Card](https://goreportcard.com/badge/github.com/akurilov/gobot)](https://goreportcard.com/report/github.com/akurilov/gobot)
[![Build Status](https://api.cirrus-ci.com/github/akurilov/gobot.svg)](https://cirrus-ci.com/github/akurilov/gobot)
# gobot
Quick and dirty internet crawler
## Build
```bash
go build cmd/gobot.go
```
## Run
```bash
./gobot <URL_TO_START_CRAWLING_FROM>
```
or
```bash
go run cmd/gobot.go <URL_TO_START_CRAWLING_FROM>
``` 
## Monitoring
The Prometheus metrics are being exposed @ `http://localhost:2112/metrics`
