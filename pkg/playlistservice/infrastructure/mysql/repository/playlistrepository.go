package repository

import (
	"database/sql"
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"playlistservice/pkg/playlistservice/domain"
	"time"
)

func NewPlaylistRepository(client mysql.Client) domain.PlaylistRepository {
	return &playlistRepository{
		client: client,
	}
}

type playlistRepository struct {
	client mysql.Client
}

func (repo *playlistRepository) NewID() domain.PlaylistID {
	return domain.PlaylistID(uuid.New())
}

func (repo *playlistRepository) Find(id domain.PlaylistID) (domain.Playlist, error) {
	const selectSql = `SELECT * from playlist WHERE playlist_id = ?`

	binaryUUID, err := uuid.UUID(id).MarshalBinary()
	if err != nil {
		return domain.Playlist{}, err
	}

	var playlist sqlxPlaylist

	err = repo.client.Get(&playlist, selectSql, binaryUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Playlist{}, domain.ErrPlaylistNotFound
		}
		return domain.Playlist{}, errors.WithStack(err)
	}

	return domain.Playlist{
		ID:        domain.PlaylistID(playlist.ID),
		Name:      playlist.Name,
		OwnerID:   domain.PlaylistOwnerID(playlist.OwnerID),
		CreatedAt: playlist.CreatedAt,
		UpdatedAt: nil,
	}, nil
}

func (repo *playlistRepository) Store(playlist domain.Playlist) error {
	const insertSql = `
		INSERT INTO playlist (playlist_id, name, owner_id, created_at, updated_at) VALUES(?, ?, ?, ?, ?)
		ON DUPLICATE KEY 
		UPDATE playlist_id=VALUES(playlist_id), name=VALUES(name), owner_id=VALUES(owner_id), created_at=VALUES(created_at), updated_at=VALUES(updated_at)
	`

	now := time.Now()
	playlist.UpdatedAt = &now

	if playlist.CreatedAt == nil {

		playlist.CreatedAt = &now
	}

	binaryUUID, err := uuid.UUID(playlist.ID).MarshalBinary()
	if err != nil {
		return errors.WithStack(err)
	}

	ownerID, err := uuid.UUID(playlist.OwnerID).MarshalBinary()
	if err != nil {
		return err
	}

	_, err = repo.client.Exec(insertSql, binaryUUID, playlist.Name, ownerID, playlist.CreatedAt, playlist.UpdatedAt)
	return err
}

func (repo *playlistRepository) Remove(id domain.PlaylistID) error {
	const deleteSql = `DELETE FROM playlist WHERE playlist_id = ?`

	binaryUUID, err := uuid.UUID(id).MarshalBinary()
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = repo.client.Exec(deleteSql, binaryUUID)
	return err
}

type sqlxPlaylist struct {
	ID        uuid.UUID
	Name      string
	OwnerID   uuid.UUID
	CreatedAt *time.Time
	UpdatedAt *time.Time
}
