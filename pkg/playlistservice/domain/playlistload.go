package domain

import (
	"time"
)

type PlaylistData interface {
	ID() PlaylistID
	Name() string
	OwnerID() PlaylistOwnerID
	Items() []PlaylistItemData
	CreatedAt() *time.Time
	UpdatedAt() *time.Time
}

type PlaylistItemData interface {
	ID() PlaylistItemID
	ContentID() ContentID
	CreatedAt() *time.Time
}

func LoadPlaylist(data PlaylistData) Playlist {
	return Playlist{
		id:        data.ID(),
		name:      data.Name(),
		ownerID:   data.OwnerID(),
		items:     mapItems(data.Items()),
		createdAt: data.CreatedAt(),
		updatedAt: data.UpdatedAt(),
	}
}

func mapItems(items []PlaylistItemData) map[PlaylistItemID]PlaylistItem {
	result := make(map[PlaylistItemID]PlaylistItem)
	for _, item := range items {
		result[item.ID()] = PlaylistItem{
			id:        item.ID(),
			contentID: item.ContentID(),
			createdAt: item.CreatedAt(),
		}
	}
	return result

}
