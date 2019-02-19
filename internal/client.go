package internal

import (
	"crypto/tls"
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
	contentLengthLimit int
}

func NewGobotClient(contentLengthLimit int) *GobotClient {
	return &GobotClient{
		http.Client{
			Transport: &http.Transport{
				IdleConnTimeout: 1 * time.Second,
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
		contentLengthLimit,
	}
}

func (client *GobotClient) ContentText(url string) (string, error) {
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
	var contentLen int
	if len(contentLenRaw) == 0 {
		contentLen = client.contentLengthLimit
	} else {
		contentLen, err = strconv.Atoi(contentLenRaw)
		if err != nil {
			return "", errors.New("failed to parse the content length header value " + contentLenRaw)
		}
		if client.contentLengthLimit < contentLen {
			contentLen = client.contentLengthLimit
		}
	}
	content := make([]byte, contentLen, contentLen)
	contentReader := resp.Body
	defer func() {
		err := contentReader.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()
	contentLen, err = io.ReadFull(contentReader, content)
	if err == io.ErrUnexpectedEOF {
		err = nil // discard
	}
	txt := string(content[:contentLen])
	return txt, err
}
