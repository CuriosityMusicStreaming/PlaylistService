package infrastructure

import (
	"context"
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	contentserviceapi "playlistservice/api/contentservice"
	"playlistservice/pkg/integrationtests/app"
)

func NewContentServiceApi(client contentserviceapi.ContentServiceClient, serializer auth.UserDescriptorSerializer) *contentServiceApi {
	return &contentServiceApi{client: client, serializer: serializer}
}

type contentServiceApi struct {
	client     contentserviceapi.ContentServiceClient
	serializer auth.UserDescriptorSerializer
}

func (api *contentServiceApi) AddContent(
	title string,
	contentType contentserviceapi.ContentType,
	availabilityType contentserviceapi.ContentAvailabilityType,
	userDescriptor auth.UserDescriptor,
) (*contentserviceapi.AddContentResponse, error) {
	userToken, err := api.serializer.Serialize(userDescriptor)
	if err != nil {
		panic(err)
	}

	resp, err := api.client.AddContent(context.Background(), &contentserviceapi.AddContentRequest{
		Name:             title,
		Type:             contentType,
		AvailabilityType: availabilityType,
		UserToken:        userToken,
	})
	return resp, api.transformError(err)
}

func (api *contentServiceApi) GetAuthorContent(userDescriptor auth.UserDescriptor) (*contentserviceapi.GetAuthorContentResponse, error) {
	userToken, err := api.serializer.Serialize(userDescriptor)
	if err != nil {
		panic(err)
	}

	resp, err := api.client.GetAuthorContent(context.Background(), &contentserviceapi.GetAuthorContentRequest{
		UserToken: userToken,
	})
	return resp, api.transformError(err)
}

func (api *contentServiceApi) GetContentList(contentIDs []string) (*contentserviceapi.GetContentListResponse, error) {
	return api.client.GetContentList(context.Background(), &contentserviceapi.GetContentListRequest{
		ContentIDs: contentIDs,
	})
}

func (api *contentServiceApi) DeleteContent(userDescriptor auth.UserDescriptor, contentID string) error {
	userToken, err := api.serializer.Serialize(userDescriptor)
	if err != nil {
		panic(err)
	}

	_, err = api.client.DeleteContent(context.Background(), &contentserviceapi.DeleteContentRequest{
		ContentID: contentID,
		UserToken: userToken,
	})
	return api.transformError(err)
}

func (api *contentServiceApi) SetContentAvailabilityType(
	userDescriptor auth.UserDescriptor,
	contentID string,
	contentAvailabilityType contentserviceapi.ContentAvailabilityType,
) error {
	userToken, err := api.serializer.Serialize(userDescriptor)
	if err != nil {
		panic(err)
	}

	_, err = api.client.SetContentAvailabilityType(context.Background(), &contentserviceapi.SetContentAvailabilityTypeRequest{
		ContentID:                  contentID,
		NewContentAvailabilityType: contentAvailabilityType,
		UserToken:                  userToken,
	})
	return api.transformError(err)
}

func (api *contentServiceApi) transformError(err error) error {
	s, ok := status.FromError(err)
	if ok {
		switch s.Code() {
		case codes.InvalidArgument:
			return app.ErrOnlyAuthorCanCreateContent
		case codes.PermissionDenied:
			return app.ErrOnlyAuthorCanManageContent
		}
	}
	return err
}
