package transport

import (
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	api "playlistservice/api/playlistservice"
	"playlistservice/pkg/playlistservice/app/query"
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

func (server *playlistServiceServer) GetPlaylist(_ context.Context, req *api.GetPlaylistRequest) (*api.GetPlaylistResponse, error) {
	userDesc, err := server.container.UserDescriptorSerializer().Deserialize(req.UserToken)
	if err != nil {
		return nil, err
	}

	queryService := server.container.PlaylistQueryService()

	playlistID, err := uuid.Parse(req.PlaylistID)
	if err != nil {
		return nil, err
	}

	playlists, err := queryService.GetPlaylists(query.PlaylistSpecification{
		OwnerIDs:    []uuid.UUID{userDesc.UserID},
		PlaylistIDs: []uuid.UUID{playlistID},
	})
	if err != nil {
		return nil, err
	}

	if len(playlists) == 0 {
		return nil, status.Errorf(codes.NotFound, "playlist not found")
	}

	playlist := playlists[0]

	return &api.GetPlaylistResponse{
		Name:               playlist.Name,
		OwnerID:            playlist.OwnerID.String(),
		CreatedAtTimestamp: uint64(playlist.CreatedAt.Unix()),
		UpdatedAtTimestamp: uint64(playlist.UpdatedAt.Unix()),
		PlaylistItems:      convertPlaylistItemViewsToApi(playlist.PlaylistItems),
	}, nil
}

func (server *playlistServiceServer) GetUserPlaylists(_ context.Context, req *api.GetUserPlaylistsRequest) (*api.GetUserPlaylistsResponse, error) {
	userDesc, err := server.container.UserDescriptorSerializer().Deserialize(req.UserToken)
	if err != nil {
		return nil, err
	}

	queryService := server.container.PlaylistQueryService()

	playlists, err := queryService.GetPlaylists(query.PlaylistSpecification{OwnerIDs: []uuid.UUID{userDesc.UserID}})
	if err != nil {
		return nil, err
	}

	result := make([]*api.Playlist, len(playlists))
	for i, playlistView := range playlists {
		result[i] = convertPlaylistViewToApi(playlistView)
	}

	return &api.GetUserPlaylistsResponse{
		Playlists: result,
	}, nil
}

func convertPlaylistViewToApi(view query.PlaylistView) *api.Playlist {
	return &api.Playlist{
		PlaylistID:         view.ID.String(),
		Name:               view.Name,
		OwnerID:            view.OwnerID.String(),
		CreatedAtTimestamp: uint64(view.CreatedAt.Unix()),
		UpdatedAtTimestamp: uint64(view.UpdatedAt.Unix()),
		PlaylistItems:      convertPlaylistItemViewsToApi(view.PlaylistItems),
	}
}

func convertPlaylistItemViewsToApi(views []query.PlaylistItemView) []*api.PlaylistItem {
	result := make([]*api.PlaylistItem, len(views))
	for i, view := range views {
		result[i] = convertPlaylistItemViewToApi(view)
	}
	return result
}

func convertPlaylistItemViewToApi(view query.PlaylistItemView) *api.PlaylistItem {
	return &api.PlaylistItem{
		PlaylistItemID:     view.ID.String(),
		ContentID:          view.ContentID.String(),
		CreatedAtTimestamp: uint64(view.CreatedAt.Unix()),
	}
}
