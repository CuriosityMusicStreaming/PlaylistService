package mysql

import (
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/infrastructure/mysql"
	"github.com/pkg/errors"
	"playlistservice/pkg/playlistservice/app/service"
	"playlistservice/pkg/playlistservice/domain"
	"playlistservice/pkg/playlistservice/infrastructure/mysql/repository"
)

func NewUnitOfFactory(client mysql.TransactionalClient) service.UnitOfWorkFactory {
	return &unitOfWorkFactory{client: client}
}

type unitOfWorkFactory struct {
	client mysql.TransactionalClient
}

func (factory *unitOfWorkFactory) NewUnitOfWork(_ string) (service.UnitOfWork, error) {
	transaction, err := factory.client.BeginTransaction()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &unitOfWork{transaction: transaction}, nil
}

type unitOfWork struct {
	transaction mysql.Transaction
}

func (u *unitOfWork) PlaylistRepository() domain.PlaylistRepository {
	return repository.NewPlaylistRepository(u.transaction)
}

func (u *unitOfWork) PlaylistItemRepository() domain.PlaylistItemRepository {
	return repository.NewPlaylistItemRepository(u.transaction)
}

func (u *unitOfWork) Complete(err error) error {
	if err != nil {
		err2 := u.transaction.Rollback()
		if err2 != nil {
			return errors.Wrap(err, err2.Error())
		}
	}

	return errors.WithStack(u.transaction.Commit())
}
