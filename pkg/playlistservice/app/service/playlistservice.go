package service

import (
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/auth"
	"github.com/google/uuid"

	"playlistservice/pkg/playlistservice/domain"
)

const (
	playlistLockName = "playlist-service-lock"
)

type PlaylistService interface {
	CreatePlaylist(name string, userDescriptor auth.UserDescriptor) (uuid.UUID, error)
	SetPlaylistName(id uuid.UUID, userDescriptor auth.UserDescriptor, newName string) error
	AddToPlaylist(id uuid.UUID, userDescriptor auth.UserDescriptor, contentID uuid.UUID) (uuid.UUID, error)
	RemoveFromPlaylist(id uuid.UUID, userDescriptor auth.UserDescriptor) error
	RemovePlaylist(id uuid.UUID, userDescriptor auth.UserDescriptor) error

	RemoveContentFromPlaylists(contentIDs []uuid.UUID, playlistIDs []uuid.UUID) error
}

func NewPlaylistService(
	contentService ContentChecker,
	unitOfWorkFactory UnitOfWorkFactory,
	eventDispatcher domain.EventDispatcher,
	remover PlaylistRemover,
) PlaylistService {
	return &playlistService{
		contentService:    contentService,
		unitOfWorkFactory: unitOfWorkFactory,
		eventDispatcher:   eventDispatcher,
		remover:           remover,
	}
}

type playlistService struct {
	contentService    ContentChecker
	unitOfWorkFactory UnitOfWorkFactory
	eventDispatcher   domain.EventDispatcher
	remover           PlaylistRemover
}

func (service *playlistService) CreatePlaylist(name string, userDescriptor auth.UserDescriptor) (uuid.UUID, error) {
	var playlistID domain.PlaylistID
	err := service.executeInUnitOfWorkWithServiceLock(playlistLockName, func(provider RepositoryProvider) error {
		domainService := service.domainPlaylistService(provider)

		var err error

		playlistID, err = domainService.CreatePlaylist(name, domain.PlaylistOwnerID(userDescriptor.UserID))

		return err
	})

	return uuid.UUID(playlistID), err
}

func (service *playlistService) SetPlaylistName(id uuid.UUID, userDescriptor auth.UserDescriptor, newName string) error {
	return service.executeInUnitOfWorkWithServiceLock(playlistLockName+id.String(), func(provider RepositoryProvider) error {
		domainService := service.domainPlaylistService(provider)

		return domainService.SetPlaylistName(domain.PlaylistID(id), domain.PlaylistOwnerID(userDescriptor.UserID), newName)
	})
}

func (service *playlistService) AddToPlaylist(id uuid.UUID, userDescriptor auth.UserDescriptor, contentID uuid.UUID) (uuid.UUID, error) {
	err := service.contentService.ContentExists([]uuid.UUID{contentID})
	if err != nil {
		return uuid.UUID{}, err
	}

	var playlistItemID domain.PlaylistItemID
	err = service.executeInUnitOfWorkWithServiceLock(playlistLockName+id.String(), func(provider RepositoryProvider) error {
		var err2 error

		playlistItemID, err2 = service.domainPlaylistService(provider).AddToPlaylist(
			domain.PlaylistID(id),
			domain.PlaylistOwnerID(userDescriptor.UserID),
			domain.ContentID(contentID),
		)
		return err2
	})

	return uuid.UUID(playlistItemID), err
}

func (service *playlistService) RemoveFromPlaylist(id uuid.UUID, userDescriptor auth.UserDescriptor) error {
	return service.executeInUnitOfWorkWithServiceLock(playlistLockName+id.String(), func(provider RepositoryProvider) error {
		return service.domainPlaylistService(provider).RemoveFromPlaylist(
			domain.PlaylistItemID(id),
			domain.PlaylistOwnerID(userDescriptor.UserID),
		)
	})
}

func (service *playlistService) RemovePlaylist(id uuid.UUID, userDescriptor auth.UserDescriptor) error {
	return service.executeInUnitOfWorkWithServiceLock(playlistLockName+id.String(), func(provider RepositoryProvider) error {
		return service.domainPlaylistService(provider).RemovePlaylist(
			domain.PlaylistID(id),
			domain.PlaylistOwnerID(userDescriptor.UserID),
		)
	})
}

func (service *playlistService) RemoveContentFromPlaylists(contentIDs []uuid.UUID, playlistIDs []uuid.UUID) error {
	domainContentIDs := make([]domain.ContentID, 0, len(contentIDs))
	for _, contentID := range contentIDs {
		domainContentIDs = append(domainContentIDs, domain.ContentID(contentID))
	}

	domainPlaylistIDs := make([]domain.PlaylistID, 0, len(playlistIDs))
	for _, playlistID := range playlistIDs {
		domainPlaylistIDs = append(domainPlaylistIDs, domain.PlaylistID(playlistID))
	}

	return service.executeInUnitOfWorkWithServiceLock("clear-playlists-contents", func(provider RepositoryProvider) error {
		return service.domainPlaylistService(provider).RemoveFromPlaylists(
			domainContentIDs,
			domainPlaylistIDs,
		)
	})
}

func (service *playlistService) executeInUnitOfWorkWithServiceLock(lockName string, f func(provider RepositoryProvider) error) error {
	return service.executeInUnitOfWork(lockName, f)
}

func (service playlistService) executeInUnitOfWork(lockName string, f func(provider RepositoryProvider) error) error {
	unitOfWork, err := service.unitOfWorkFactory.NewUnitOfWork(lockName)
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
	return domain.NewPlaylistService(provider.PlaylistRepository(), service.eventDispatcher)
}
