package query

import (
	"github.com/google/uuid"
	"time"
)

type PlaylistView struct {
	ID            uuid.UUID
	Name          string
	OwnerID       uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
	PlaylistItems []PlaylistItemView
}

type PlaylistItemView struct {
	ID        uuid.UUID
	ContentID uuid.UUID
	CreatedAt *time.Time
}

type PlaylistSpecification struct {
	PlaylistIDs []uuid.UUID
	OwnerIDs    []uuid.UUID
}

type PlaylistQueryService interface {
	GetPlaylists(spec PlaylistSpecification) ([]PlaylistView, error)
}
