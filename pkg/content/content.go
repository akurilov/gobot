package content

import (
	"net/url"
)

type Content struct {
	MimeType string
	Source   *url.URL
	Body     []byte
}

func NewContent(mimeType string, source *url.URL, body []byte) *Content {
	return &Content{
		MimeType: mimeType,
		Source:   source,
		Body:     body,
	}
}

var (
	emptyMimeType = ""
	emptyBody     = make([]byte, 0)
)

func NewDummyContent(source *url.URL) *Content {
	return &Content{
		MimeType: emptyMimeType,
		Source:   source,
		Body:     emptyBody,
	}
}

func (c *Content) IsDummy() bool {
	return c.MimeType == emptyMimeType && len(c.Body) == 0
}
