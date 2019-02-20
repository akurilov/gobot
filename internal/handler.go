package internal

import (
	"github.com/akurilov/gobot/pkg/content"
	"github.com/akurilov/gobot/pkg/content/handle"
	"go.uber.org/zap"
)

type Handler struct {
	log          *zap.Logger
	contentQueue <-chan *content.Content
	delegates    []handle.ContentHandle
}

func NewHandler(log *zap.Logger, contentQueue <-chan *content.Content, delegates []handle.ContentHandle) *Handler {
	return &Handler{
		log,
		contentQueue,
		delegates,
	}
}

func (h *Handler) Loop() {
	for c := range h.contentQueue {
		for _, handler := range h.delegates {
			err := handler.Handle(c)
			if err != nil {
				h.log.Warn(err.Error())
			}
		}
	}
}
