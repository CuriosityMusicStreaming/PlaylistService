package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"playlistservice/pkg/playlistservice/domain"
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
	const selectSQL = `SELECT * from playlist WHERE playlist_id = ?`

	binaryUUID, err := uuid.UUID(id).MarshalBinary()
	if err != nil {
		return domain.Playlist{}, errors.WithStack(err)
	}

	var playlist sqlxPlaylist

	err = repo.client.Get(&playlist, selectSQL, binaryUUID)
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

func (repo *playlistRepository) FindAll(ids []domain.PlaylistID) ([]domain.Playlist, error) {
	const selectSQL = `SELECT * from playlist WHERE playlist_id IN (?)`

	binaryUUIDs := make([][]byte, 0, len(ids))

	for _, id := range ids {
		bytes, err := uuid.UUID(id).MarshalBinary()
		if err != nil {
			return nil, err
		}

		binaryUUIDs = append(binaryUUIDs, bytes)
	}

	var playlists []sqlxPlaylist

	query, args, err := sqlx.In(selectSQL, ids)
	if err != nil {
		return nil, err
	}

	err = repo.client.Get(&playlists, query, args)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.WithStack(err)
	}

	playlistsItemsMap, err := repo.fetchPlaylistItemsMap(binaryUUIDs)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Playlist, 0, len(playlists))

	for _, playlist := range playlists {
		result = append(result, domain.LoadPlaylist(&playlistData{
			id:        playlist.ID,
			name:      playlist.Name,
			ownerID:   playlist.OwnerID,
			items:     convertPlaylistItems(playlistsItemsMap[playlist.ID]),
			createdAt: playlist.CreatedAt,
			updatedAt: playlist.UpdatedAt,
		}))
	}

	return result, nil
}

func (repo *playlistRepository) FindByItemID(playlistItemID domain.PlaylistItemID) (domain.Playlist, error) {
	const selectSQL = `
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

	binaryUUID, err := uuid.UUID(playlistItemID).MarshalBinary()
	if err != nil {
		return domain.Playlist{}, err
	}

	var playlist sqlxPlaylist

	err = repo.client.Get(&playlist, selectSQL, binaryUUID)
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
	const insertSQL = `
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

	_, err = repo.client.Exec(insertSQL, binaryUUID, playlist.Name(), ownerID, playlist.CreatedAt(), playlist.UpdatedAt())
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
	const deleteSQL = `DELETE FROM playlist WHERE playlist_id = ?`

	binaryUUID, err := uuid.UUID(id).MarshalBinary()
	if err != nil {
		return errors.WithStack(err)
	}

	err = repo.removePlaylistItems(id)
	if err != nil {
		return err
	}

	_, err = repo.client.Exec(deleteSQL, binaryUUID)
	if err != nil {
		return err
	}

	return nil
}

func (repo *playlistRepository) fetchPlaylistItems(id uuid.UUID) ([]sqlxPlaylistItem, error) {
	const selectSQL = `SELECT playlist_item_id, playlist_id, content_id, created_at from playlist_item WHERE playlist_id = ?`

	binaryUUID, err := id.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var playlistItems []sqlxPlaylistItem

	err = repo.client.Select(&playlistItems, selectSQL, binaryUUID)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return playlistItems, nil
}

func (repo *playlistRepository) fetchPlaylistItemsMap(ids [][]byte) (map[uuid.UUID][]sqlxPlaylistItem, error) {
	const selectSQL = `SELECT playlist_item_id, playlist_id, content_id, created_at from playlist_item WHERE playlist_id IN (?)`

	var playlistsItems []sqlxPlaylistItem

	query, args, err := sqlx.In(selectSQL, ids)
	if err != nil {
		return nil, err
	}

	err = repo.client.Select(&playlistsItems, query, args)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := map[uuid.UUID][]sqlxPlaylistItem{}

	for _, playlistItem := range playlistsItems {
		items, ok := result[playlistItem.PlaylistID]
		if !ok {
			items = []sqlxPlaylistItem{}
		}
		items = append(items, playlistItem)
		result[playlistItem.PlaylistID] = items
	}

	return result, nil
}

//nolint
func (repo *playlistRepository) storePlaylistItems(playlistID domain.PlaylistID, items map[domain.PlaylistItemID]domain.PlaylistItem) error {
	if len(items) == 0 {
		return nil
	}

	const insertSQL = `
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

	_, err := repo.client.Exec(fmt.Sprintf(insertSQL, strings.Join(values, ", ")), args...)
	return errors.WithStack(err)
}

func (repo *playlistRepository) removeDeletedItems(playlistID domain.PlaylistID, items map[domain.PlaylistItemID]domain.PlaylistItem) error {
	deleteSQL := `DELETE FROM playlist_item WHERE playlist_id = ?`

	binaryPlaylistID, err := uuid.UUID(playlistID).MarshalBinary()
	if err != nil {
		return errors.WithStack(err)
	}

	args := []interface{}{binaryPlaylistID}

	if len(items) != 0 {
		playlistItemIDs := make([][]byte, 0, len(items))
		for itemID := range items {
			binary, err2 := uuid.UUID(itemID).MarshalBinary()
			if err2 != nil {
				return err2
			}
			playlistItemIDs = append(playlistItemIDs, binary)
		}

		query, params, err2 := sqlx.In("AND playlist_item_id NOT IN (?)", playlistItemIDs)
		if err2 != nil {
			return err2
		}

		deleteSQL += " " + query
		args = append(args, params...)
	}

	_, err = repo.client.Exec(deleteSQL, args...)
	return err
}

func (repo playlistRepository) removePlaylistItems(playlistID domain.PlaylistID) error {
	const deleteSQL = `DELETE FROM playlist_item WHERE playlist_id = ?`

	id, err := uuid.UUID(playlistID).MarshalBinary()
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = repo.client.Exec(deleteSQL, id)
	return err
}

func convertPlaylistItems(sqlxItems []sqlxPlaylistItem) []domain.PlaylistItemData {
	result := make([]domain.PlaylistItemData, 0, len(sqlxItems))
	for _, item := range sqlxItems {
		result = append(result, &playlistItemData{
			id:        item.ID,
			contentID: item.ContentID,
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
	ID         uuid.UUID  `db:"playlist_item_id"`
	PlaylistID uuid.UUID  `db:"playlist_id"`
	ContentID  uuid.UUID  `db:"content_id"`
	CreatedAt  *time.Time `db:"created_at"`
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
