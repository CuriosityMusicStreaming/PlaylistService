package service

import (
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"playlistservice/pkg/playlistservice/app/service"
)

func NewPlaylistRemover(client mysql.Client) service.PlaylistRemover {
	return &playlistRemover{client: client}
}

type playlistRemover struct {
	client mysql.Client
}

func (remover *playlistRemover) RemoveFromPlaylists(contentIDs []uuid.UUID) error {
	ids := convertIDs(contentIDs)
	query, args, err := sqlx.In(`DELETE FROM playlist_item WHERE content_id IN (?)`, ids)
	if err != nil {
		return err
	}

	_, err = remover.client.Exec(query, args...)
	return err
}

func convertIDs(ids []uuid.UUID) [][]byte {
	result := make([][]byte, 0, len(ids))
	for _, id := range ids {
		binaryID, _ := id.MarshalBinary()
		result = append(result, binaryID)
	}
	return result
}
