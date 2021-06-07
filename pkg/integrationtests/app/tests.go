package app

import (
	"context"
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/auth"
	playlistserviceapi "playlistservice/api/playlistservice"
)

type UserContainer interface {
	AddAuthor(descriptor auth.UserDescriptor)
	AddListener(descriptor auth.UserDescriptor)

	Clear()
}

func RunTests(playlistServiceClient playlistserviceapi.PlayListServiceClient, container UserContainer) {
	playlistTests(&playlistServiceApiFacade{
		client:     playlistServiceClient,
		serializer: auth.NewUserDescriptorSerializer(),
	},
		container,
	)
}

type playlistServiceApiFacade struct {
	client     playlistserviceapi.PlayListServiceClient
	serializer auth.UserDescriptorSerializer
}

func (facade *playlistServiceApiFacade) CreatePlaylist(title string, userDescriptor auth.UserDescriptor) (string, error) {
	userToken, err := facade.serializer.Serialize(userDescriptor)
	assertNoErr(err)

	resp, err := facade.client.CreatePlaylist(context.Background(), &playlistserviceapi.CreatePlaylistRequest{
		Name:      title,
		UserToken: userToken,
	})
	if err != nil {
		return "", err
	}

	return resp.PlaylistID, nil
}

func (facade *playlistServiceApiFacade) GetPlaylist(playlistID string, userDescriptor auth.UserDescriptor) (*playlistserviceapi.GetPlaylistResponse, error) {
	userToken, err := facade.serializer.Serialize(userDescriptor)
	assertNoErr(err)

	return facade.client.GetPlaylist(context.Background(), &playlistserviceapi.GetPlaylistRequest{
		PlaylistID: playlistID,
		UserToken:  userToken,
	})
}

func (facade *playlistServiceApiFacade) GetUserPlaylists(userDescriptor auth.UserDescriptor) (*playlistserviceapi.GetUserPlaylistsResponse, error) {
	userToken, err := facade.serializer.Serialize(userDescriptor)
	assertNoErr(err)

	return facade.client.GetUserPlaylists(context.Background(), &playlistserviceapi.GetUserPlaylistsRequest{
		UserToken: userToken,
	})
}

func (facade *playlistServiceApiFacade) DeletePlaylist(playlistID string, userDescriptor auth.UserDescriptor) error {
	userToken, err := facade.serializer.Serialize(userDescriptor)
	assertNoErr(err)

	_, err = facade.client.RemovePlaylist(context.Background(), &playlistserviceapi.RemovePlaylistRequest{
		PlaylistID: playlistID,
		UserToken:  userToken,
	})
	return err
}
