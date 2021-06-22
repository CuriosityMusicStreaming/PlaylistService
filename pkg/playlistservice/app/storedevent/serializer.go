package storedevent

import (
	"encoding/json"

	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/storedevent"
	commondomain "github.com/CuriosityMusicStreaming/ComponentsPool/pkg/domain"
	"github.com/google/uuid"

	"playlistservice/pkg/playlistservice/domain"
)

func NewEventSerializer() storedevent.EventSerializer {
	return &eventSerializer{}
}

type eventSerializer struct {
}

type eventBody struct {
	Type    string
	Payload *json.RawMessage
}

func (serializer *eventSerializer) Serialize(event commondomain.Event) (string, error) {
	payload, err := serializeAsJSON(event)
	if err != nil {
		return "", err
	}

	payloadRawMessage := json.RawMessage(payload)
	body := eventBody{
		Type:    event.ID(),
		Payload: &payloadRawMessage,
	}

	messageBody, err := json.Marshal(body)

	return string(messageBody), err
}

func serializeAsJSON(event commondomain.Event) ([]byte, error) {
	return json.Marshal(serializeEvent(event))
}

//nolint
func serializeEvent(event commondomain.Event) (eventPayload interface{}) {
	switch currEvent := event.(type) {
	case domain.PlaylistCreated:
		eventPayload = struct {
			PlaylistID uuid.UUID `json:"playlist_id"`
			OwnerID    uuid.UUID `json:"owner_id"`
		}{
			PlaylistID: uuid.UUID(currEvent.PlaylistID),
			OwnerID:    uuid.UUID(currEvent.OwnerID),
		}
	case domain.PlaylistNameChanged:
		eventPayload = struct {
			PlaylistID uuid.UUID `json:"playlist_id"`
			Name       string    `json:"name"`
		}{
			PlaylistID: uuid.UUID(currEvent.PlaylistID),
			Name:       currEvent.NewName,
		}
	case domain.PlaylistItemAdded:
		eventPayload = struct {
			PlaylistID     uuid.UUID `json:"playlist_id"`
			PlaylistItemID uuid.UUID `json:"playlist_item_id"`
			ContentID      uuid.UUID `json:"content_id"`
		}{
			PlaylistID:     uuid.UUID(currEvent.PlaylistID),
			PlaylistItemID: uuid.UUID(currEvent.PlaylistItemID),
			ContentID:      uuid.UUID(currEvent.ContentID),
		}
	case domain.PlaylistItemRemoved:
		eventPayload = struct {
			PlaylistID     uuid.UUID `json:"playlist_id"`
			PlaylistItemID uuid.UUID `json:"playlist_item_id"`
		}{
			PlaylistID:     uuid.UUID(currEvent.PlaylistID),
			PlaylistItemID: uuid.UUID(currEvent.PlaylistItemID),
		}
	case domain.PlaylistRemoved:
		eventPayload = struct {
			PlaylistID uuid.UUID `json:"playlist_id"`
			OwnerID    uuid.UUID `json:"owner_id"`
		}{
			PlaylistID: uuid.UUID(currEvent.PlaylistID),
			OwnerID:    uuid.UUID(currEvent.OwnerID),
		}
	case domain.PlaylistItemsRemoved:
		eventPayload = struct {
			PlaylistIDToPlaylistItemIDsMap map[uuid.UUID][]uuid.UUID `json:"playlist_id_to_playlist_item_ids"`
		}{
			PlaylistIDToPlaylistItemIDsMap: convertPlaylistIDToPlaylistItemIDsMapToUUIDsMap(currEvent.PlaylistIDToPlaylistItemIDsMap),
		}
	}
	return
}

func convertPlaylistIDToPlaylistItemIDsMapToUUIDsMap(idsMap map[domain.PlaylistID][]domain.PlaylistItemID) map[uuid.UUID][]uuid.UUID {
	result := map[uuid.UUID][]uuid.UUID{}
	for playlistID, playlistItemIDs := range idsMap {
		convertedPlaylistItemIDs := make([]uuid.UUID, 0, len(playlistItemIDs))
		for _, playlistItemID := range playlistItemIDs {
			convertedPlaylistItemIDs = append(convertedPlaylistItemIDs, uuid.UUID(playlistItemID))
		}
		result[uuid.UUID(playlistID)] = convertedPlaylistItemIDs
	}
	return result
}
