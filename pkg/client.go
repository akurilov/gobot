package pkg

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type GobotClient struct {
	http.Client
}

func NewGobotClient() *GobotClient {
	return &GobotClient{
		http.Client{
			Transport: &http.Transport{
				MaxIdleConns:    200,
				IdleConnTimeout: 10 * time.Second,
			},
		},
	}
}

func (client *GobotClient) ContentText(url string, maxLength int) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	statusCode := resp.StatusCode
	if statusCode < 200 || statusCode > 299 {
		return "", errors.New("response status code " + strconv.Itoa(statusCode))
	}
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text") {
		return "", errors.New("unsupported content type " + contentType)
	}
	contentLenRaw := resp.Header.Get("Content-Length")
	if len(contentLenRaw) == 0 {
		return "", errors.New("missing content length header")
	}
	contentLen, err := strconv.Atoi(contentLenRaw)
	if err != nil {
		return "", errors.New("failed to parse the content length header value " + contentLenRaw)
	}
	contentLen = MinInt(contentLen, maxLength)
	content := make([]byte, 0, contentLen)
	contentReader := resp.Body
	defer func() {
		err := contentReader.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()
	contentLen, err = io.ReadFull(contentReader, content)
	txt := string(content[:contentLen])
	return txt, err
}
