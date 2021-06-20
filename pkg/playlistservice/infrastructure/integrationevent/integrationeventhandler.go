package integrationevent

import (
	"encoding/json"
	"fmt"

	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/logger"
	"github.com/google/uuid"

	"playlistservice/pkg/playlistservice/app/service"
)

const (
	privateContentAvailabilityType int = 1
)

type DependencyContainer interface {
	PlaylistService() service.PlaylistService
}

func NewIntegrationEventHandler(logger logger.Logger, container DependencyContainer) Handler {
	return &integrationEventHandler{
		logger:    logger,
		container: container,
	}
}

type integrationEventHandler struct {
	logger    logger.Logger
	container DependencyContainer
}

func (handler *integrationEventHandler) Handle(msgBody string) error {
	var e event

	err := json.Unmarshal([]byte(msgBody), &e)
	if err != nil {
		handler.logger.Error(err, fmt.Sprintf("Failed to unmarshall integration event with body %s", msgBody))
		return err
	}

	handler.logger.Info(fmt.Sprintf("Integration event received with body %s", msgBody))

	err = handler.handleEvents(e)
	if err != nil {
		handler.logger.Error(err, fmt.Sprintf("Failed to handle integration event with type %s", e.Type))
		return err
	}

	return nil
}

func (handler *integrationEventHandler) handleEvents(e event) error {
	if e.Type == "content_availability_type_changed" {
		payload := contentAvailabilityTypeChangedPayload{}
		err := json.Unmarshal(e.Payload, &payload)
		if err != nil {
			return err
		}

		if payload.ContentAvailabilityType != privateContentAvailabilityType {
			return nil
		}

		contentID, err := uuid.Parse(payload.ContentID)
		if err != nil {
			return err
		}

		return handler.container.PlaylistService().RemoveFromPlaylists([]uuid.UUID{contentID})
	}

	if e.Type == "content_deleted" {
		payload := contentDeletedPayload{}
		err := json.Unmarshal(e.Payload, &payload)
		if err != nil {
			return err
		}

		contentID, err := uuid.Parse(payload.ContentID)
		if err != nil {
			return err
		}

		return handler.container.PlaylistService().RemoveFromPlaylists([]uuid.UUID{contentID})
	}

	return nil
}

type event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type contentAvailabilityTypeChangedPayload struct {
	ContentID               string `json:"content_id"`
	ContentAvailabilityType int    `json:"content_availability_type"`
}

type contentDeletedPayload struct {
	ContentID string `json:"content_id"`
}
