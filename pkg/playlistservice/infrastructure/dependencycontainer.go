package infrastructure

import (
	commonauth "github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/auth"
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/logger"
	commonmysql "github.com/CuriosityMusicStreaming/ComponentsPool/pkg/infrastructure/mysql"
	"playlistservice/pkg/playlistservice/app/query"
	"playlistservice/pkg/playlistservice/app/service"
	"playlistservice/pkg/playlistservice/domain"
	"playlistservice/pkg/playlistservice/infrastructure/mysql"
	mysqlquery "playlistservice/pkg/playlistservice/infrastructure/mysql/query"
)

type DependencyContainer interface {
	PlaylistService() service.PlaylistService
	PlaylistQueryService() query.PlaylistQueryService
	UserDescriptorSerializer() commonauth.UserDescriptorSerializer
}

func NewDependencyContainer(
	client commonmysql.TransactionalClient,
	logger logger.Logger,
) DependencyContainer {
	return &dependencyContainer{
		client:            client,
		logger:            logger,
		eventDispatcher:   eventDispatcher(logger),
		unitOfWorkFactory: unitOfWorkFactory(client),
	}
}

type dependencyContainer struct {
	client            commonmysql.TransactionalClient
	logger            logger.Logger
	eventDispatcher   domain.EventDispatcher
	unitOfWorkFactory service.UnitOfWorkFactory
}

func (container *dependencyContainer) PlaylistService() service.PlaylistService {
	return service.NewPlaylistService(container.unitOfWorkFactory, container.eventDispatcher)
}

func (container *dependencyContainer) PlaylistQueryService() query.PlaylistQueryService {
	return mysqlquery.NewPlaylistQueryService(container.client)
}

func (container *dependencyContainer) UserDescriptorSerializer() commonauth.UserDescriptorSerializer {
	return commonauth.NewUserDescriptorSerializer()
}

func unitOfWorkFactory(client commonmysql.TransactionalClient) service.UnitOfWorkFactory {
	return mysql.NewUnitOfFactory(client)
}

func eventDispatcher(logger logger.Logger) domain.EventDispatcher {
	eventPublisher := domain.NewEventPublisher()

	{

	}

	return eventPublisher
}
