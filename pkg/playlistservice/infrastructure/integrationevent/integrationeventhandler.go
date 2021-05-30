package integrationevent

import (
	"fmt"
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/logger"
)

func NewIntegrationEventHandler(logger logger.Logger) Handler {
	return &integrationEventListener{logger: logger}
}

type integrationEventListener struct {
	logger logger.Logger
}

func (handler *integrationEventListener) Handle(msgBody string) error {
	handler.logger.Info(fmt.Sprintf("Event received with body %s", msgBody))

	return nil
}
