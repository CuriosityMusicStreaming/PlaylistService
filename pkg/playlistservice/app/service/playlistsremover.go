package service

import "github.com/google/uuid"

type PlaylistRemover interface {
	RemoveFromPlaylists(contentIDs []uuid.UUID) error
}
