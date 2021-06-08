package app

import (
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/auth"
	"github.com/pkg/errors"
	contentserviceapi "playlistservice/api/contentservice"
	playlistserviceapi "playlistservice/api/playlistservice"
)

type UserContainer interface {
	AddAuthor(descriptor auth.UserDescriptor)
	AddListener(descriptor auth.UserDescriptor)

	Clear()
}

func RunTests(playlistServiceApi PlaylistServiceApi, contentServiceApi ContentServiceApi, container UserContainer) {
	playlistTests(playlistServiceApi, container)
	playlistsContentTests(playlistServiceApi, contentServiceApi, container)
}

type PlaylistServiceApi interface {
	CreatePlaylist(title string, userDescriptor auth.UserDescriptor) (string, error)
	GetPlaylist(playlistID string, userDescriptor auth.UserDescriptor) (*playlistserviceapi.GetPlaylistResponse, error)
	GetUserPlaylists(userDescriptor auth.UserDescriptor) (*playlistserviceapi.GetUserPlaylistsResponse, error)
	SetPlaylistTitle(playlistID string, title string, userDescriptor auth.UserDescriptor) error
	DeletePlaylist(playlistID string, userDescriptor auth.UserDescriptor) error

	AddToPlaylist(playlistID string, contentID string, userDescriptor auth.UserDescriptor) (string, error)
	RemoveFromPlaylist(playlistItemID string, userDescriptor auth.UserDescriptor) error
}

type ContentServiceApi interface {
	AddContent(title string, contentType contentserviceapi.ContentType, availabilityType contentserviceapi.ContentAvailabilityType, userDescriptor auth.UserDescriptor) (*contentserviceapi.AddContentResponse, error)
	GetAuthorContent(userDescriptor auth.UserDescriptor) (*contentserviceapi.GetAuthorContentResponse, error)
	GetContentList(contentIDs []string) (*contentserviceapi.GetContentListResponse, error)
	DeleteContent(userDescriptor auth.UserDescriptor, contentID string) error
	SetContentAvailabilityType(userDescriptor auth.UserDescriptor, contentID string, contentAvailabilityType contentserviceapi.ContentAvailabilityType) error
}

var (
	ErrOnlyOwnerCanManagePlaylist = errors.New("only owner can manage playlist")
	ErrContentNotFound            = errors.New("content not found")

	ErrOnlyAuthorCanCreateContent = errors.New("only author can create content")
	ErrOnlyAuthorCanManageContent = errors.New("only author can manage content")
)
