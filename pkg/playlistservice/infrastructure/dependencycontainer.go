package infrastructure

import (
	commonauth "github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/auth"
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/logger"
	commonmysql "github.com/CuriosityMusicStreaming/ComponentsPool/pkg/infrastructure/mysql"
	contentserviceapi "playlistservice/api/contentservice"
	"playlistservice/pkg/playlistservice/app/query"
	"playlistservice/pkg/playlistservice/app/service"
	"playlistservice/pkg/playlistservice/domain"
	"playlistservice/pkg/playlistservice/infrastructure/mysql"
	mysqlquery "playlistservice/pkg/playlistservice/infrastructure/mysql/query"
	infrastructureservice "playlistservice/pkg/playlistservice/infrastructure/service"
)

type DependencyContainer interface {
	PlaylistService() service.PlaylistService
	PlaylistQueryService() query.PlaylistQueryService
	UserDescriptorSerializer() commonauth.UserDescriptorSerializer
}

func NewDependencyContainer(
	client commonmysql.TransactionalClient,
	logger logger.Logger,
	contentServiceClient contentserviceapi.ContentServiceClient,
) DependencyContainer {
	return &dependencyContainer{
		playlistService: playlistService(
			contentChecker(contentServiceClient),
			unitOfWorkFactory(client),
			eventDispatcher(logger),
		),
		playlistQueryService:     playlistQueryService(client),
		userDescriptorSerializer: userDescriptorSerializer(),
	}
}

type dependencyContainer struct {
	playlistService          service.PlaylistService
	playlistQueryService     query.PlaylistQueryService
	userDescriptorSerializer commonauth.UserDescriptorSerializer
}

func (container *dependencyContainer) PlaylistService() service.PlaylistService {
	return container.playlistService
}

func (container *dependencyContainer) PlaylistQueryService() query.PlaylistQueryService {
	return container.playlistQueryService
}

func (container *dependencyContainer) UserDescriptorSerializer() commonauth.UserDescriptorSerializer {
	return container.userDescriptorSerializer
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

func playlistService(
	contentChecker service.ContentChecker,
	unitOfWork service.UnitOfWorkFactory,
	eventDispatcher domain.EventDispatcher,
) service.PlaylistService {
	return service.NewPlaylistService(
		contentChecker,
		unitOfWork,
		eventDispatcher,
	)
}

func playlistQueryService(client commonmysql.TransactionalClient) query.PlaylistQueryService {
	return mysqlquery.NewPlaylistQueryService(client)
}

func userDescriptorSerializer() commonauth.UserDescriptorSerializer {
	return commonauth.NewUserDescriptorSerializer()
}

func contentChecker(contentServiceClient contentserviceapi.ContentServiceClient) service.ContentChecker {
	return infrastructureservice.NewContentChecker(contentServiceClient)
}
