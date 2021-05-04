package service

import (
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/auth"
	"github.com/google/uuid"
	"playlistservice/pkg/playlistservice/domain"
)

type PlaylistService interface {
	CreatePlaylist(name string, userDescriptor auth.UserDescriptor) (uuid.UUID, error)
	SetPlaylistName(id uuid.UUID, userDescriptor auth.UserDescriptor, newName string) error
	AddToPlaylist(id uuid.UUID, userDescriptor auth.UserDescriptor, contentID uuid.UUID) (uuid.UUID, error)
	RemoveFromPlaylist(id uuid.UUID, userDescriptor auth.UserDescriptor) error
	RemovePlaylist(id uuid.UUID, userDescriptor auth.UserDescriptor) error
}

func NewPlaylistService(unitOfWorkFactory UnitOfWorkFactory, eventDispatcher domain.EventDispatcher) PlaylistService {
	return &playlistService{
		unitOfWorkFactory: unitOfWorkFactory,
		eventDispatcher:   eventDispatcher,
	}
}

type playlistService struct {
	unitOfWorkFactory UnitOfWorkFactory
	eventDispatcher   domain.EventDispatcher
}

func (service *playlistService) CreatePlaylist(name string, userDescriptor auth.UserDescriptor) (uuid.UUID, error) {
	var playlistID domain.PlaylistID
	err := service.executeInUnitOfWork(func(provider RepositoryProvider) error {
		domainService := domain.NewPlaylistService(
			provider.PlaylistRepository(),
			provider.PlaylistItemRepository(),
			service.eventDispatcher,
		)

		var err error

		playlistID, err = domainService.CreatePlaylist(name, domain.PlaylistOwnerID(userDescriptor.UserID))

		return err
	})

	return uuid.UUID(playlistID), err
}

func (service *playlistService) SetPlaylistName(id uuid.UUID, userDescriptor auth.UserDescriptor, newName string) error {
	return service.executeInUnitOfWork(func(provider RepositoryProvider) error {
		domainService := domain.NewPlaylistService(
			provider.PlaylistRepository(),
			provider.PlaylistItemRepository(),
			service.eventDispatcher,
		)

		return domainService.SetPlaylistName(domain.PlaylistID(id), domain.PlaylistOwnerID(userDescriptor.UserID), newName)
	})
}

func (service *playlistService) AddToPlaylist(id uuid.UUID, userDescriptor auth.UserDescriptor, contentID uuid.UUID) (uuid.UUID, error) {
	var playlistItemID domain.PlaylistItemID
	err := service.executeInUnitOfWork(func(provider RepositoryProvider) error {
		var err error
		playlistItemID, err = service.domainPlaylistService(provider).AddToPlaylist(
			domain.PlaylistID(id),
			domain.PlaylistOwnerID(userDescriptor.UserID),
			domain.ContentID(contentID),
		)
		return err
	})

	return uuid.UUID(playlistItemID), err
}

func (service *playlistService) RemoveFromPlaylist(id uuid.UUID, userDescriptor auth.UserDescriptor) error {
	return service.executeInUnitOfWork(func(provider RepositoryProvider) error {
		return service.domainPlaylistService(provider).RemoveFromPlaylist(
			domain.PlaylistItemID(id),
			domain.PlaylistOwnerID(userDescriptor.UserID),
		)
	})
}

func (service *playlistService) RemovePlaylist(id uuid.UUID, userDescriptor auth.UserDescriptor) error {
	return service.executeInUnitOfWork(func(provider RepositoryProvider) error {
		return service.domainPlaylistService(provider).RemovePlaylist(
			domain.PlaylistID(id),
			domain.PlaylistOwnerID(userDescriptor.UserID),
		)
	})
}

func (service *playlistService) executeInUnitOfWork(f func(provider RepositoryProvider) error) error {
	unitOfWork, err := service.unitOfWorkFactory.NewUnitOfWork("")
	if err != nil {
		return err
	}
	defer func() {
		err = unitOfWork.Complete(err)
	}()
	err = f(unitOfWork)
	return err
}

func (service *playlistService) domainPlaylistService(provider RepositoryProvider) domain.PlaylistService {
	return domain.NewPlaylistService(
		provider.PlaylistRepository(),
		provider.PlaylistItemRepository(),
		service.eventDispatcher,
	)
}
