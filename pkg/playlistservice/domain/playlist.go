package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmptyPlaylistName      = errors.New("empty playlist name")
	ErrPlaylistNotFound       = errors.New("playlist not found")
	ErrPlaylistItemNotFound   = errors.New("playlist item not found")
	ErrPlaylistByItemNotFound = errors.New("playlist by item not found")
)

type (
	PlaylistID      uuid.UUID
	PlaylistItemID  uuid.UUID
	PlaylistOwnerID uuid.UUID
	ContentID       uuid.UUID
)

func NewPlaylist(id PlaylistID, name string, ownerID PlaylistOwnerID) (Playlist, error) {
	if name == "" {
		return Playlist{}, ErrEmptyPlaylistName
	}

	now := time.Now()

	return Playlist{
		id:        id,
		name:      name,
		ownerID:   ownerID,
		items:     map[PlaylistItemID]PlaylistItem{},
		createdAt: &now,
		updatedAt: &now,
	}, nil
}

type Playlist struct {
	id        PlaylistID
	name      string
	ownerID   PlaylistOwnerID
	items     map[PlaylistItemID]PlaylistItem
	createdAt *time.Time
	updatedAt *time.Time
}

func (playlist *Playlist) ID() PlaylistID {
	return playlist.id
}

func (playlist *Playlist) Name() string {
	return playlist.name
}

func (playlist *Playlist) SetName(newName string) {
	playlist.name = newName
}

func (playlist *Playlist) OwnerID() PlaylistOwnerID {
	return playlist.ownerID
}

func (playlist *Playlist) CreatedAt() *time.Time {
	return playlist.createdAt
}

func (playlist *Playlist) UpdatedAt() *time.Time {
	return playlist.updatedAt
}

func (playlist *Playlist) Items() map[PlaylistItemID]PlaylistItem {
	return playlist.items
}

func (playlist *Playlist) AddItem(id PlaylistItemID, contentID ContentID) {
	playlistItem, ok := playlist.items[id]
	if ok {
		playlistItem.contentID = contentID
		return
	}

	now := time.Now()

	playlist.items[id] = PlaylistItem{
		id:        id,
		contentID: contentID,
		createdAt: &now,
	}
	playlist.updatedAt = &now
}

func (playlist *Playlist) RemoveItem(itemID PlaylistItemID) error {
	_, exists := playlist.items[itemID]
	if !exists {
		return ErrPlaylistItemNotFound
	}

	delete(playlist.items, itemID)

	now := time.Now()
	playlist.updatedAt = &now

	return nil
}

func (playlist *Playlist) RemoveContent(contentIDs []ContentID) ([]PlaylistItemID, error) {
	var removedPlaylistItemIDs []PlaylistItemID
	for id, item := range playlist.items {
		for _, contentID := range contentIDs {
			if item.contentID == contentID {
				removedPlaylistItemIDs = append(removedPlaylistItemIDs, id)
				delete(playlist.items, id)
			}
		}
	}
	return removedPlaylistItemIDs, nil
}

type PlaylistItem struct {
	id        PlaylistItemID
	contentID ContentID
	createdAt *time.Time
}

func (item *PlaylistItem) ID() PlaylistItemID {
	return item.id
}

func (item *PlaylistItem) ContentID() ContentID {
	return item.contentID
}

func (item *PlaylistItem) CreatedAt() *time.Time {
	return item.createdAt
}

type PlaylistSpecification struct {
	ContentIDs []ContentID
}

type PlaylistRepository interface {
	NewID() PlaylistID
	NewPlaylistItemID() PlaylistItemID
	Find(id PlaylistID) (Playlist, error)
	FindByItemID(playlistItemID PlaylistItemID) (Playlist, error)
	FindAll(ids []PlaylistID) ([]Playlist, error)
	Store(playlist Playlist) error
	Remove(id PlaylistID) error
}
