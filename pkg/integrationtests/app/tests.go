package app

import (
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/auth"
	"github.com/pkg/errors"
	playlistserviceapi "playlistservice/api/playlistservice"
)

type UserContainer interface {
	AddAuthor(descriptor auth.UserDescriptor)
	AddListener(descriptor auth.UserDescriptor)

	Clear()
}

func RunTests(playlistServiceApi PlaylistServiceApi, container UserContainer) {
	playlistTests(playlistServiceApi, container)
}

type PlaylistServiceApi interface {
	CreatePlaylist(title string, userDescriptor auth.UserDescriptor) (string, error)
	GetPlaylist(playlistID string, userDescriptor auth.UserDescriptor) (*playlistserviceapi.GetPlaylistResponse, error)
	GetUserPlaylists(userDescriptor auth.UserDescriptor) (*playlistserviceapi.GetUserPlaylistsResponse, error)
	SetPlaylistTitle(playlistID string, title string, userDescriptor auth.UserDescriptor) error
	DeletePlaylist(playlistID string, userDescriptor auth.UserDescriptor) error
}

var (
	ErrOnlyOwnerCanManagePlaylist = errors.New("only owner can manage playlist")
)
