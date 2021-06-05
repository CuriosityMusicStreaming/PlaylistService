package repository

import (
	"database/sql"
	"fmt"
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"playlistservice/pkg/playlistservice/domain"
	"strings"
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

func (repo *playlistRepository) NewPlaylistItemID() domain.PlaylistItemID {
	return domain.PlaylistItemID(uuid.New())
}

func (repo *playlistRepository) Find(id domain.PlaylistID) (domain.Playlist, error) {
	const selectSql = `SELECT * from playlist WHERE playlist_id = ?`

	binaryUUID, err := uuid.UUID(id).MarshalBinary()
	if err != nil {
		return domain.Playlist{}, errors.WithStack(err)
	}

	var playlist sqlxPlaylist

	err = repo.client.Get(&playlist, selectSql, binaryUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Playlist{}, domain.ErrPlaylistNotFound
		}
		return domain.Playlist{}, errors.WithStack(err)
	}

	playlistItems, err := repo.fetchPlaylistItems(playlist.ID)
	if err != nil {
		return domain.Playlist{}, errors.WithStack(err)
	}

	return domain.LoadPlaylist(&playlistData{
		id:        playlist.ID,
		name:      playlist.Name,
		ownerID:   playlist.OwnerID,
		items:     convertPlaylistItems(playlistItems),
		createdAt: playlist.CreatedAt,
		updatedAt: playlist.UpdatedAt,
	}), nil
}

func (repo *playlistRepository) FindByItemID(playlistItemId domain.PlaylistItemID) (domain.Playlist, error) {
	const selectSql = `
		SELECT 
			p.playlist_id AS playlist_id, 
			p.name AS name, 
			p.owner_id AS owner_id, 
			p.created_at AS created_at, 
			p.updated_at AS updated_at
		FROM 
			playlist p 
		LEFT JOIN playlist_item pi on p.playlist_id = pi.playlist_id 
		WHERE pi.playlist_item_id = ?
	`

	binaryUUID, err := uuid.UUID(playlistItemId).MarshalBinary()
	if err != nil {
		return domain.Playlist{}, err
	}

	var playlist sqlxPlaylist

	err = repo.client.Get(&playlist, selectSql, binaryUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Playlist{}, domain.ErrPlaylistByItemNotFound
		}
		return domain.Playlist{}, errors.WithStack(err)
	}

	playlistItems, err := repo.fetchPlaylistItems(playlist.ID)
	if err != nil {
		return domain.Playlist{}, err
	}

	return domain.LoadPlaylist(&playlistData{
		id:        playlist.ID,
		name:      playlist.Name,
		ownerID:   playlist.OwnerID,
		items:     convertPlaylistItems(playlistItems),
		createdAt: playlist.CreatedAt,
		updatedAt: playlist.UpdatedAt,
	}), nil
}

func (repo *playlistRepository) Store(playlist domain.Playlist) error {
	const insertSql = `
		INSERT INTO playlist (playlist_id, name, owner_id, created_at, updated_at) VALUES(?, ?, ?, ?, ?)
		ON DUPLICATE KEY 
		UPDATE playlist_id=VALUES(playlist_id), name=VALUES(name), owner_id=VALUES(owner_id), created_at=VALUES(created_at), updated_at=VALUES(updated_at)
	`

	binaryUUID, err := uuid.UUID(playlist.ID()).MarshalBinary()
	if err != nil {
		return errors.WithStack(err)
	}

	ownerID, err := uuid.UUID(playlist.OwnerID()).MarshalBinary()
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = repo.client.Exec(insertSql, binaryUUID, playlist.Name(), ownerID, playlist.CreatedAt(), playlist.UpdatedAt())
	if err != nil {
		return errors.WithStack(err)
	}

	err = repo.storePlaylistItems(playlist.ID(), playlist.Items())
	if err != nil {
		return errors.WithStack(err)
	}

	err = repo.removeDeletedItems(playlist.ID(), playlist.Items())
	if err != nil {
		return errors.WithStack(err)
	}

	return err
}

func (repo *playlistRepository) Remove(id domain.PlaylistID) error {
	const deleteSql = `DELETE FROM playlist WHERE playlist_id = ?`

	binaryUUID, err := uuid.UUID(id).MarshalBinary()
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = repo.client.Exec(deleteSql, binaryUUID)
	if err != nil {
		return err
	}

	return repo.removePlaylistItems(id)
}

func (repo *playlistRepository) fetchPlaylistItems(id uuid.UUID) ([]sqlxPlaylistItem, error) {
	const selectSql = `SELECT playlist_item_id, content_id, created_at from playlist_item WHERE playlist_id = ?`

	binaryUUID, err := id.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var playlistItems []sqlxPlaylistItem

	err = repo.client.Select(&playlistItems, selectSql, binaryUUID)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return playlistItems, nil
}

func (repo *playlistRepository) storePlaylistItems(playlistID domain.PlaylistID, items map[domain.PlaylistItemID]domain.PlaylistItem) error {
	if len(items) == 0 {
		return nil
	}

	const insertSql = `
		INSERT INTO playlist_item (playlist_item_id, playlist_id, content_id, created_at) VALUES %s
		ON DUPLICATE KEY 
		UPDATE playlist_item_id=VALUES(playlist_item_id), playlist_id=VALUES(playlist_id), content_id=VALUES(content_id), created_at=VALUES(created_at)
	`

	values := make([]string, 0, len(items))
	var args []interface{}

	for _, item := range items {
		itemID, err := uuid.UUID(item.ID()).MarshalBinary()
		if err != nil {
			return errors.WithStack(err)
		}
		args = append(args, itemID)

		binaryPlaylistID, err := uuid.UUID(playlistID).MarshalBinary()
		if err != nil {
			return errors.WithStack(err)
		}
		args = append(args, binaryPlaylistID)

		contentID, err := uuid.UUID(item.ContentID()).MarshalBinary()
		if err != nil {
			return errors.WithStack(err)
		}
		args = append(args, contentID)

		args = append(args, item.CreatedAt())

		values = append(values, "(?, ?, ?, ?)")
	}

	_, err := repo.client.Exec(fmt.Sprintf(insertSql, strings.Join(values, ", ")), args...)
	return errors.WithStack(err)
}

func (repo *playlistRepository) removeDeletedItems(playlistID domain.PlaylistID, items map[domain.PlaylistItemID]domain.PlaylistItem) error {
	if len(items) == 0 {
		return nil
	}

	const deleteSql = `DELETE FROM playlist_item WHERE playlist_item_id NOT IN (?) AND playlist_id = ?`

	binaryPlaylistID, err := uuid.UUID(playlistID).MarshalBinary()
	if err != nil {
		return errors.WithStack(err)
	}

	playlistItemIDs := make([][]byte, 0, len(items))
	for itemID := range items {
		binary, err2 := uuid.UUID(itemID).MarshalBinary()
		if err2 != nil {
			return err2
		}
		playlistItemIDs = append(playlistItemIDs, binary)
	}

	query, args, err := sqlx.In(deleteSql, playlistItemIDs, binaryPlaylistID)
	if err != nil {
		return err
	}

	_, err = repo.client.Exec(query, args...)
	return err
}

func (repo playlistRepository) removePlaylistItems(playlistID domain.PlaylistID) error {
	const deleteSql = `DELETE FROM playlist_item WHERE playlist_id = ?`

	id, err := uuid.UUID(playlistID).MarshalBinary()
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = repo.client.Exec(deleteSql, id)
	return err
}

func convertPlaylistItems(sqlxItems []sqlxPlaylistItem) []domain.PlaylistItemData {
	result := make([]domain.PlaylistItemData, 0, len(sqlxItems))
	for _, item := range sqlxItems {
		result = append(result, &playlistItemData{
			id:        item.ID,
			contentID: item.ID,
			createdAt: item.CreatedAt,
		})
	}
	return result
}

type sqlxPlaylist struct {
	ID        uuid.UUID  `db:"playlist_id"`
	Name      string     `db:"name"`
	OwnerID   uuid.UUID  `db:"owner_id"`
	CreatedAt *time.Time `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}

type sqlxPlaylistItem struct {
	ID        uuid.UUID  `db:"playlist_item_id"`
	ContentID uuid.UUID  `db:"content_id"`
	CreatedAt *time.Time `db:"created_at"`
}

type playlistData struct {
	id        uuid.UUID
	name      string
	ownerID   uuid.UUID
	items     []domain.PlaylistItemData
	createdAt *time.Time
	updatedAt *time.Time
}

func (p *playlistData) ID() domain.PlaylistID {
	return domain.PlaylistID(p.id)
}

func (p *playlistData) Name() string {
	return p.name
}

func (p *playlistData) OwnerID() domain.PlaylistOwnerID {
	return domain.PlaylistOwnerID(p.ownerID)
}

func (p *playlistData) CreatedAt() *time.Time {
	return p.createdAt
}

func (p *playlistData) Items() []domain.PlaylistItemData {
	return p.items
}

func (p *playlistData) UpdatedAt() *time.Time {
	return p.updatedAt
}

type playlistItemData struct {
	id        uuid.UUID
	contentID uuid.UUID
	createdAt *time.Time
}

func (p *playlistItemData) ID() domain.PlaylistItemID {
	return domain.PlaylistItemID(p.id)
}

func (p *playlistItemData) ContentID() domain.ContentID {
	return domain.ContentID(p.contentID)
}

func (p *playlistItemData) CreatedAt() *time.Time {
	return p.createdAt
}
