package transport

import (
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"playlistservice/pkg/playlistservice/domain"
)

func translateError(err error) error {
	switch errors.Cause(err) {
	case domain.ErrPlaylistItemNotFound:
	case domain.ErrPlaylistNotFound:
		return status.Error(codes.NotFound, err.Error())
	case domain.ErrOnlyOwnerCanManagePlaylist:
		return status.Error(codes.PermissionDenied, err.Error())
	}

	return err
}
