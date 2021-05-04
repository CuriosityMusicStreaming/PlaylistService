package service

import "github.com/google/uuid"

type ContentService interface {
	GetContent(contentIDs []uuid.UUID)
}
