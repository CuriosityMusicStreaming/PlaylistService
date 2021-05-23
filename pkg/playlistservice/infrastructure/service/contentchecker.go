package service

import (
	"context"
	"github.com/google/uuid"
	contentserviceapi "playlistservice/api/contentservice"
	"playlistservice/pkg/playlistservice/app/service"
)

func NewContentChecker(contentServiceClient contentserviceapi.ContentServiceClient) service.ContentChecker {
	return &contentChecker{contentServiceClient: contentServiceClient}
}

type contentChecker struct {
	contentServiceClient contentserviceapi.ContentServiceClient
}

func (checker *contentChecker) ContentExists(contentIDs []uuid.UUID) error {
	ctx := context.Background()
	resp, err := checker.contentServiceClient.GetContentList(ctx, &contentserviceapi.GetContentListRequest{
		ContentIDs: uuidsToStrings(contentIDs),
	})
	if err != nil {
		return err
	}

	if len(resp.Contents) == 0 {
		return service.ErrContentNotFound
	}

	return nil
}

func uuidsToStrings(ids []uuid.UUID) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		result = append(result, id.String())
	}
	return result
}
