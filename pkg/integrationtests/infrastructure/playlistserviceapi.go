package infrastructure

import (
	"context"

	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	playlistserviceapi "playlistservice/api/playlistservice"
	"playlistservice/pkg/integrationtests/app"
)

func NewPlaylistServiceAPI(
	client playlistserviceapi.PlayListServiceClient,
	serializer auth.UserDescriptorSerializer,
) app.PlaylistServiceAPI {
	return &playlistServiceAPI{
		client:     client,
		serializer: serializer,
	}
}

type playlistServiceAPI struct {
	client     playlistserviceapi.PlayListServiceClient
	serializer auth.UserDescriptorSerializer
}

func (api *playlistServiceAPI) CreatePlaylist(title string, userDescriptor auth.UserDescriptor) (string, error) {
	userToken, err := api.serializer.Serialize(userDescriptor)
	if err != nil {
		panic(err)
	}

	resp, err := api.client.CreatePlaylist(context.Background(), &playlistserviceapi.CreatePlaylistRequest{
		Name:      title,
		UserToken: userToken,
	})
	if err != nil {
		return "", api.transformError(err)
	}

	return resp.PlaylistID, nil
}

func (api *playlistServiceAPI) GetPlaylist(playlistID string, userDescriptor auth.UserDescriptor) (*playlistserviceapi.GetPlaylistResponse, error) {
	userToken, err := api.serializer.Serialize(userDescriptor)
	if err != nil {
		panic(err)
	}

	response, err := api.client.GetPlaylist(context.Background(), &playlistserviceapi.GetPlaylistRequest{
		PlaylistID: playlistID,
		UserToken:  userToken,
	})
	return response, api.transformError(err)
}

func (api *playlistServiceAPI) GetUserPlaylists(userDescriptor auth.UserDescriptor) (*playlistserviceapi.GetUserPlaylistsResponse, error) {
	userToken, err := api.serializer.Serialize(userDescriptor)
	if err != nil {
		panic(err)
	}

	response, err := api.client.GetUserPlaylists(context.Background(), &playlistserviceapi.GetUserPlaylistsRequest{
		UserToken: userToken,
	})
	return response, api.transformError(err)
}

//nolint:gocritic
func (api *playlistServiceAPI) SetPlaylistTitle(playlistID string, title string, userDescriptor auth.UserDescriptor) error {
	userToken, err := api.serializer.Serialize(userDescriptor)
	if err != nil {
		panic(err)
	}

	_, err = api.client.SetPlaylistName(context.Background(), &playlistserviceapi.SetPlaylistNameRequest{
		PlaylistID: playlistID,
		NewName:    title,
		UserToken:  userToken,
	})

	return api.transformError(err)
}

func (api *playlistServiceAPI) DeletePlaylist(playlistID string, userDescriptor auth.UserDescriptor) error {
	userToken, err := api.serializer.Serialize(userDescriptor)
	if err != nil {
		panic(err)
	}

	_, err = api.client.RemovePlaylist(context.Background(), &playlistserviceapi.RemovePlaylistRequest{
		PlaylistID: playlistID,
		UserToken:  userToken,
	})
	return api.transformError(err)
}

//nolint:gocritic
func (api *playlistServiceAPI) AddToPlaylist(playlistID string, contentID string, userDescriptor auth.UserDescriptor) (string, error) {
	userToken, err := api.serializer.Serialize(userDescriptor)
	if err != nil {
		panic(err)
	}

	resp, err := api.client.AddToPlaylist(context.Background(), &playlistserviceapi.AddToPlaylistRequest{
		PlaylistID: playlistID,
		UserToken:  userToken,
		ContentID:  contentID,
	})
	if err != nil {
		return "", api.transformError(err)
	}

	return resp.PlaylistItemID, nil
}

func (api *playlistServiceAPI) RemoveFromPlaylist(playlistItemID string, userDescriptor auth.UserDescriptor) error {
	userToken, err := api.serializer.Serialize(userDescriptor)
	if err != nil {
		panic(err)
	}

	_, err = api.client.RemoveFromPlaylist(context.Background(), &playlistserviceapi.RemoveFromPlaylistRequest{
		PlaylistItemID: playlistItemID,
		UserToken:      userToken,
	})

	return api.transformError(err)
}

func (api *playlistServiceAPI) transformError(err error) error {
	s, ok := status.FromError(err)
	if ok {
		switch s.Code() {
		case codes.InvalidArgument:
			return app.ErrContentNotFound
		case codes.PermissionDenied:
			return app.ErrOnlyOwnerCanManagePlaylist
		}
	}
	return err
}
