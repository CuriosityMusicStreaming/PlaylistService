package repository

import (
	"database/sql"
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"playlistservice/pkg/playlistservice/domain"
	"time"
)

func NewPlaylistItemRepository(client mysql.Client) domain.PlaylistItemRepository {
	return &playlistItemRepository{client: client}
}

type playlistItemRepository struct {
	client mysql.Client
}

func (repo *playlistItemRepository) NewID() domain.PlaylistItemID {
	return domain.PlaylistItemID(uuid.New())
}

func (repo *playlistItemRepository) Find(id domain.PlaylistItemID) (domain.PlaylistItem, error) {
	const selectSql = `SELECT * from playlist_item WHERE playlist_item_id = ?`

	binaryUUID, err := uuid.UUID(id).MarshalBinary()
	if err != nil {
		return domain.PlaylistItem{}, err
	}

	var playlistItem sqlxPlaylistItem

	err = repo.client.Get(&playlistItem, selectSql, binaryUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.PlaylistItem{}, domain.ErrPlaylistItemNotFound
		}
		return domain.PlaylistItem{}, errors.WithStack(err)
	}

	return domain.PlaylistItem{
		ID:         domain.PlaylistItemID(playlistItem.ID),
		PlaylistID: domain.PlaylistID(playlistItem.PlaylistID),
		ContentID:  domain.ContentID(playlistItem.ContentID),
		CreatedAt:  playlistItem.CreatedAt,
	}, err
}

func (repo *playlistItemRepository) Store(playlistItem domain.PlaylistItem) error {
	const insertSql = `
		INSERT INTO playlist_item (playlist_item_id, playlist_id, content_id, created_at) VALUES(?, ?, ?, ?)
		ON DUPLICATE KEY 
		UPDATE playlist_item_id=VALUES(playlist_item_id), playlist_id=VALUES(playlist_id), content_id=VALUES(content_id), created_at=VALUES(created_at)
	`

	if playlistItem.CreatedAt == nil {
		now := time.Now()
		playlistItem.CreatedAt = &now
	}

	playlistItemID, err := uuid.UUID(playlistItem.ID).MarshalBinary()
	if err != nil {
		return errors.WithStack(err)
	}

	playlistID, err := uuid.UUID(playlistItem.ID).MarshalBinary()
	if err != nil {
		return errors.WithStack(err)
	}

	contentID, err := uuid.UUID(playlistItem.ID).MarshalBinary()
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = repo.client.Exec(insertSql, playlistItemID, playlistID, contentID, playlistItem.CreatedAt)
	return err
}

func (repo *playlistItemRepository) Remove(id domain.PlaylistItemID) error {
	const deleteSql = `DELETE FROM playlist_item WHERE playlist_item_id = ?`

	binaryUUID, err := uuid.UUID(id).MarshalBinary()
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = repo.client.Exec(deleteSql, binaryUUID)
	return err
}

type sqlxPlaylistItem struct {
	ID         uuid.UUID
	PlaylistID uuid.UUID
	ContentID  uuid.UUID
	CreatedAt  *time.Time
}
