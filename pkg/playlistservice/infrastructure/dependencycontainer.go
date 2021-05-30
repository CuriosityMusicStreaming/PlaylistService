package infrastructure

import (
	commonauth "github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/auth"
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/logger"
	commonstoredevent "github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/storedevent"
	commonmysql "github.com/CuriosityMusicStreaming/ComponentsPool/pkg/infrastructure/mysql"
	contentserviceapi "playlistservice/api/contentservice"
	"playlistservice/pkg/playlistservice/app/query"
	"playlistservice/pkg/playlistservice/app/service"
	"playlistservice/pkg/playlistservice/app/storedevent"
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
	eventStore commonstoredevent.Store,
	storedEventSenderCallback mysql.UnitOfWorkCompleteNotifier,
) DependencyContainer {
	unitOfWorkFactory, completeNotifier := unitOfWorkFactory(client)

	completeNotifier.subscribe(storedEventSenderCallback)

	return &dependencyContainer{
		playlistService: playlistService(
			contentChecker(contentServiceClient),
			unitOfWorkFactory,
			eventDispatcher(eventStore),
		),
		playlistQueryService:     playlistQueryService(client),
		userDescriptorSerializer: userDescriptorSerializer(),
	}
}

type completeNotifier struct {
	subscribers []mysql.UnitOfWorkCompleteNotifier
}

func (notifier *completeNotifier) subscribe(subscriber mysql.UnitOfWorkCompleteNotifier) {
	notifier.subscribers = append(notifier.subscribers, subscriber)
}

func (notifier *completeNotifier) onComplete() {
	for _, subscriber := range notifier.subscribers {
		subscriber()
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

func unitOfWorkFactory(client commonmysql.TransactionalClient) (service.UnitOfWorkFactory, *completeNotifier) {
	notifier := &completeNotifier{}

	return mysql.NewNotifyingUnitOfWorkFactory(
		mysql.NewUnitOfFactory(client),
		notifier.onComplete,
	), notifier
}

func eventDispatcher(store commonstoredevent.Store) domain.EventDispatcher {
	eventPublisher := domain.NewEventPublisher()

	{
		handler := commonstoredevent.NewStoredDomainEventHandler(store, storedevent.NewEventSerializer())
		eventPublisher.Subscribe(domain.HandlerFunc(func(event domain.Event) error {
			return handler.Handle(event)
		}))
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
