package service

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	ErrContentNotFound = errors.New("content not found")
)

type ContentChecker interface {
	ContentExists(contentIDs []uuid.UUID) error
}
