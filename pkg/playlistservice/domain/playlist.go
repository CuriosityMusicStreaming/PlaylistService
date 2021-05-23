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

func NewPlaylist(id PlaylistID, name string, ownerID PlaylistOwnerID) Playlist {
	return Playlist{
		ID:      id,
		Name:    name,
		OwnerID: ownerID,
		Items:   map[PlaylistItemID]PlaylistItem{},
	}
}

type Playlist struct {
	ID        PlaylistID
	Name      string
	OwnerID   PlaylistOwnerID
	Items     map[PlaylistItemID]PlaylistItem
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func (playlist *Playlist) AddItem(item PlaylistItem) {
	playlistItem, ok := playlist.Items[item.ID]
	if ok {
		playlistItem.ContentID = item.ContentID
		return
	}
	playlist.Items[item.ID] = item

	now := time.Now()
	playlist.UpdatedAt = &now
}

func (playlist *Playlist) RemoveItem(itemID PlaylistItemID) error {
	_, exists := playlist.Items[itemID]
	if !exists {
		return ErrPlaylistItemNotFound
	}

	delete(playlist.Items, itemID)

	now := time.Now()
	playlist.UpdatedAt = &now

	return nil
}

type PlaylistItem struct {
	ID        PlaylistItemID
	ContentID ContentID
	CreatedAt *time.Time
}

var (
	ErrPlaylistNotFound     = errors.New("playlist not found")
	ErrPlaylistItemNotFound = errors.New("playlist item not found")
)

type PlaylistRepository interface {
	NewID() PlaylistID
	NewPlaylistItemID() PlaylistItemID
	Find(id PlaylistID) (Playlist, error)
	FindByItemID(playlistItemId PlaylistItemID) (Playlist, error)
	Store(playlist Playlist) error
	Remove(id PlaylistID) error
}
