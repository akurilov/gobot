package internal

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/akurilov/gobot/pkg/content"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// GobotClient HTTP client with custom init params
type GobotClient struct {
	http.Client
	contentLengthLimit int
}

// NewGobotClient creates a new HTTP client with the specified max content length being fetched
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

// GetContent gets the content text from the specified URL. Returns error if:
// * failed to fetch
// * response status is not 2xx
// * response Content-Length header contains the value which couldn't be parsed as integer
func (client *GobotClient) GetContent(u *url.URL) (*content.Content, error) {
	resp, err := client.Get(u.String())
	if err != nil {
		return content.NewDummyContent(u), err
	}
	statusCode := resp.StatusCode
	if statusCode < 200 || statusCode > 299 {
		return content.NewDummyContent(u), errors.New("response status code " + strconv.Itoa(statusCode))
	}
	mimeType := resp.Header.Get("Content-Type")
	contentLenRaw := resp.Header.Get("Content-Length")
	var contentLen int
	if len(contentLenRaw) == 0 {
		contentLen = client.contentLengthLimit
	} else {
		contentLen, err = strconv.Atoi(contentLenRaw)
		if err != nil {
			return content.NewDummyContent(u),
				errors.New("failed to parse the content length header value " + contentLenRaw)
		}
		if client.contentLengthLimit < contentLen {
			contentLen = client.contentLengthLimit
		}
	}
	body := make([]byte, contentLen, contentLen)
	contentReader := resp.Body
	defer func() {
		err := contentReader.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()
	contentLen, err = io.ReadFull(contentReader, body)
	if err == io.ErrUnexpectedEOF {
		err = nil // discard
	}
	return content.NewContent(mimeType, u, body), err
}
