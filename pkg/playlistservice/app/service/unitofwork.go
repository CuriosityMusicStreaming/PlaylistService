package service

import "playlistservice/pkg/playlistservice/domain"

type UnitOfWorkFactory interface {
	NewUnitOfWork(lockName string) (UnitOfWork, error)
}

type RepositoryProvider interface {
	PlaylistRepository() domain.PlaylistRepository
	PlaylistItemRepository() domain.PlaylistItemRepository
}

type UnitOfWork interface {
	RepositoryProvider
	Complete(err error) error
}
