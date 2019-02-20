package handle

import (
	"github.com/akurilov/gobot/pkg/content"
)

type ContentHandle interface {
	Handle(c *content.Content) error
}
