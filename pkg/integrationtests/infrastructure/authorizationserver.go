package infrastructure

import (
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/auth"
	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"playlistservice/api/authorizationservice"
)

func NewAuthorizationServer() (authorizationservice.AuthorizationServiceServer, *UsersContainer) {
	container := &UsersContainer{}

	return &authorizationServer{
		container:                container,
		userDescriptorSerializer: auth.NewUserDescriptorSerializer(),
	}, container
}

type authorizationServer struct {
	container                *UsersContainer
	userDescriptorSerializer auth.UserDescriptorSerializer
}

func (server *authorizationServer) CanAddContent(_ context.Context, req *authorizationservice.CanAddContentRequest) (*authorizationservice.CanAddContentResponse, error) {
	userDescriptor, err := server.userDescriptorSerializer.Deserialize(req.UserToken)
	if err != nil {
		return nil, err
	}

	for _, user := range server.container.Authors {
		if user.UserID == userDescriptor.UserID {
			return &authorizationservice.CanAddContentResponse{CanAdd: true}, nil
		}
	}

	for _, user := range server.container.Listeners {
		if user.UserID == userDescriptor.UserID {
			return &authorizationservice.CanAddContentResponse{CanAdd: false}, errors.New("user can`t add content")
		}
	}

	return nil, errors.New("user not found")
}

type UsersContainer struct {
	Authors   []auth.UserDescriptor
	Listeners []auth.UserDescriptor
}

func (container *UsersContainer) AddAuthor(descriptor auth.UserDescriptor) {
	container.Authors = append(container.Authors, descriptor)
}

func (container *UsersContainer) AddListener(descriptor auth.UserDescriptor) {
	container.Listeners = append(container.Listeners, descriptor)
}

func (container *UsersContainer) Clear() {
	container.Authors = nil
	container.Listeners = nil
}
