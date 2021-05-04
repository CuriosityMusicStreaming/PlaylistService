package transport

import (
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/types/known/emptypb"
	api "playlistservice/api/playlistservice"
	"playlistservice/pkg/playlistservice/infrastructure"
)

func NewPlaylistServiceServer(container infrastructure.DependencyContainer) api.PlayListServiceServer {
	return &playlistServiceServer{
		container: container,
	}
}

type playlistServiceServer struct {
	container infrastructure.DependencyContainer
}

func (server *playlistServiceServer) CreatePlaylist(_ context.Context, req *api.CreatePlaylistRequest) (*api.CreatePlaylistResponse, error) {
	userDesc, err := server.container.UserDescriptorSerializer().Deserialize(req.UserToken)
	if err != nil {
		return nil, err
	}

	playlistService := server.container.PlaylistService()

	playlistID, err := playlistService.CreatePlaylist(req.Name, userDesc)
	if err != nil {
		return nil, err
	}

	return &api.CreatePlaylistResponse{PlaylistID: playlistID.String()}, nil
}

func (server *playlistServiceServer) AddToPlaylist(_ context.Context, req *api.AddToPlaylistRequest) (*api.AddToPlaylistResponse, error) {
	userDesc, err := server.container.UserDescriptorSerializer().Deserialize(req.UserToken)
	if err != nil {
		return nil, err
	}

	playlistService := server.container.PlaylistService()

	playlistID, err := uuid.Parse(req.PlaylistID)
	if err != nil {
		return nil, err
	}

	contentID, err := uuid.Parse(req.ContentID)
	if err != nil {
		return nil, err
	}

	playlistItemID, err := playlistService.AddToPlaylist(playlistID, userDesc, contentID)
	if err != nil {
		return nil, err
	}

	return &api.AddToPlaylistResponse{
		PlaylistItemID: playlistItemID.String(),
	}, nil
}

func (server *playlistServiceServer) RemoveFromPlaylist(_ context.Context, req *api.RemoveFromPlaylistRequest) (*emptypb.Empty, error) {
	userDesc, err := server.container.UserDescriptorSerializer().Deserialize(req.UserToken)
	if err != nil {
		return nil, err
	}

	playlistService := server.container.PlaylistService()

	playlistItemID, err := uuid.Parse(req.PlaylistItemID)
	if err != nil {
		return nil, err
	}

	err = playlistService.RemoveFromPlaylist(playlistItemID, userDesc)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (server *playlistServiceServer) RemovePlaylist(_ context.Context, req *api.RemovePlaylistRequest) (*emptypb.Empty, error) {
	userDesc, err := server.container.UserDescriptorSerializer().Deserialize(req.UserToken)
	if err != nil {
		return nil, err
	}

	playlistService := server.container.PlaylistService()

	playlistID, err := uuid.Parse(req.PlaylistID)

	err = playlistService.RemovePlaylist(playlistID, userDesc)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (server *playlistServiceServer) GetPlaylist(ctx context.Context, req *api.GetPlaylistRequest) (*api.GetPlaylistResponse, error) {
	return nil, nil
}
