package app

import (
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
