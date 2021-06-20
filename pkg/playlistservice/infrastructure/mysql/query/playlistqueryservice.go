package query

import (
	"fmt"
	"strings"
	"time"

	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/infrastructure/mysql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"playlistservice/pkg/playlistservice/app/query"
)

func NewPlaylistQueryService(client mysql.Client) query.PlaylistQueryService {
	return &playlistQueryService{client: client}
}

type playlistQueryService struct {
	client mysql.Client
}

func (service *playlistQueryService) GetPlaylists(spec query.PlaylistSpecification) ([]query.PlaylistView, error) {
	selectSql := `SELECT * FROM playlist`

	conditions, args, err := getWhereConditionsBySpec(spec)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if conditions != "" {
		selectSql += fmt.Sprintf(` WHERE %s`, conditions)
	}

	var playlists []sqlxPlaylistView

	err = service.client.Select(&playlists, selectSql, args...)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(playlists) == 0 {
		return nil, errors.WithStack(err)
	}

	playlistsIDs := make([]uuid.UUID, len(playlists))
	for i, playlist := range playlists {
		playlistsIDs[i] = playlist.ID
	}

	playlistsItemsMap, err := service.getPlaylistsItemsMap(playlistsIDs)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := make([]query.PlaylistView, len(playlists))

	for i, playlist := range playlists {
		result[i] = query.PlaylistView{
			ID:        playlist.ID,
			Name:      playlist.Name,
			OwnerID:   playlist.OwnerID,
			CreatedAt: playlist.CreatedAt,
			UpdatedAt: playlist.UpdatedAt,
		}
		items, ok := playlistsItemsMap[playlist.ID]
		if !ok {
			continue
		}
		result[i].PlaylistItems = convertToPlaylistItemViews(items)
	}

	return result, nil
}

func (service *playlistQueryService) getPlaylistsItemsMap(playlistIDs []uuid.UUID) (map[uuid.UUID][]sqlxPlaylistItemView, error) {
	playlistsItems, err := service.getPlaylistsItems(playlistIDs)
	if err != nil {
		return nil, err
	}

	result := map[uuid.UUID][]sqlxPlaylistItemView{}

	for _, item := range playlistsItems {
		items, ok := result[item.PlaylistID]
		if !ok {
			items = []sqlxPlaylistItemView{}
		}
		items = append(items, item)
		result[item.PlaylistID] = items
	}

	return result, nil
}

func (service *playlistQueryService) getPlaylistsItems(playlistIDs []uuid.UUID) ([]sqlxPlaylistItemView, error) {
	ids, err := uuidsToBinaryUUIDs(playlistIDs)
	if err != nil {
		return nil, err
	}

	sqlQuery, args, err := sqlx.In(`SELECT * FROM playlist_item WHERE playlist_id IN (?)`, ids)
	if err != nil {
		return nil, err
	}

	var playlistItems []sqlxPlaylistItemView

	err = service.client.Select(&playlistItems, sqlQuery, args...)

	return playlistItems, err
}

func getWhereConditionsBySpec(spec query.PlaylistSpecification) (string, []interface{}, error) {
	var conditions []string
	var params []interface{}

	if len(spec.OwnerIDs) != 0 {
		ids, err := uuidsToBinaryUUIDs(spec.OwnerIDs)
		if err != nil {
			return "", nil, errors.WithStack(err)
		}
		sqlQuery, args, err := sqlx.In(`owner_id IN (?)`, ids)
		if err != nil {
			return "", nil, errors.WithStack(err)
		}
		conditions = append(conditions, sqlQuery)
		for _, arg := range args {
			params = append(params, arg)
		}
	}

	if len(spec.PlaylistIDs) != 0 {
		ids, err := uuidsToBinaryUUIDs(spec.PlaylistIDs)
		if err != nil {
			return "", nil, errors.WithStack(err)
		}
		sqlQuery, args, err := sqlx.In(`playlist_id IN (?)`, ids)
		if err != nil {
			return "", nil, errors.WithStack(err)
		}
		conditions = append(conditions, sqlQuery)
		for _, arg := range args {
			params = append(params, arg)
		}
	}

	return strings.Join(conditions, " AND "), params, nil
}

func uuidsToBinaryUUIDs(uuids []uuid.UUID) ([][]byte, error) {
	res := make([][]byte, len(uuids))
	for i, id := range uuids {
		binaryUUID, err := id.MarshalBinary()
		if err != nil {
			return nil, err
		}
		res[i] = binaryUUID
	}
	return res, nil
}

func convertToPlaylistItemViews(views []sqlxPlaylistItemView) []query.PlaylistItemView {
	result := make([]query.PlaylistItemView, len(views))
	for i, view := range views {
		result[i] = convertToPlaylistItemView(view)
	}
	return result
}

func convertToPlaylistItemView(view sqlxPlaylistItemView) query.PlaylistItemView {
	return query.PlaylistItemView{
		ID:        view.ID,
		ContentID: view.ContentID,
		CreatedAt: view.CreatedAt,
	}
}

type sqlxPlaylistView struct {
	ID        uuid.UUID `db:"playlist_id"`
	Name      string    `db:"name"`
	OwnerID   uuid.UUID `db:"owner_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type sqlxPlaylistItemView struct {
	ID         uuid.UUID  `db:"playlist_item_id"`
	PlaylistID uuid.UUID  `db:"playlist_id"`
	ContentID  uuid.UUID  `db:"content_id"`
	CreatedAt  *time.Time `db:"created_at"`
}
