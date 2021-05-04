package domain

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

type (
	PlaylistID      uuid.UUID
	PlaylistItemID  uuid.UUID
	PlaylistOwnerID uuid.UUID
	ContentID       uuid.UUID
)

type Playlist struct {
	ID        PlaylistID
	Name      string
	OwnerID   PlaylistOwnerID
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

type PlaylistItem struct {
	ID         PlaylistItemID
	PlaylistID PlaylistID
	ContentID  ContentID
	CreatedAt  *time.Time
}

var (
	ErrPlaylistNotFound     = errors.New("playlist not found")
	ErrPlaylistItemNotFound = errors.New("playlist item not found")
)

type PlaylistRepository interface {
	NewID() PlaylistID
	Find(id PlaylistID) (Playlist, error)
	Store(playlist Playlist) error
	Remove(id PlaylistID) error
}

type PlaylistItemRepository interface {
	NewID() PlaylistItemID
	Find(id PlaylistItemID) (PlaylistItem, error)
	Store(playlistItem PlaylistItem) error
	Remove(id PlaylistItemID) error
}
