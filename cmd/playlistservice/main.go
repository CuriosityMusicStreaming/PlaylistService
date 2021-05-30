package main

import (
	"context"
	log "github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/logger"
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/storedevent"
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/infrastructure/amqp"
	jsonlog "github.com/CuriosityMusicStreaming/ComponentsPool/pkg/infrastructure/logger"
	commonmysql "github.com/CuriosityMusicStreaming/ComponentsPool/pkg/infrastructure/mysql"
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/infrastructure/server"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"io"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	contentserviceapi "playlistservice/api/contentservice"
	"playlistservice/api/playlistservice"
	migrationsembedder "playlistservice/data/mysql"
	"playlistservice/pkg/playlistservice/infrastructure"
	"playlistservice/pkg/playlistservice/infrastructure/integrationevent"
	"playlistservice/pkg/playlistservice/infrastructure/mysql"
	"playlistservice/pkg/playlistservice/infrastructure/transport"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

var appID = "UNKNOWN"

func main() {
	logger, err := initLogger()
	if err != nil {
		stdlog.Fatal("failed to initialize logger")
	}

	config, err := parseEnv()
	if err != nil {
		logger.FatalError(err)
	}

	err = runService(config, logger)
	if err == server.ErrStopped {
		logger.Info("service is successfully stopped")
	} else if err != nil {
		logger.FatalError(err)
	}
}

func runService(config *config, logger log.MainLogger) error {
	dsn := commonmysql.DSN{
		User:     config.DatabaseUser,
		Password: config.DatabasePassword,
		Host:     config.DatabaseHost,
		Database: config.DatabaseName,
	}
	connector := commonmysql.NewConnector()
	err := connector.MigrateUp(dsn, migrationsembedder.MigrationsEmbedder)
	if err != nil {
		logger.Error(err, "failed to migrate")
	}
	err = connector.Open(dsn, config.MaxDatabaseConnections)
	if err != nil {
		return err
	}
	defer connector.Close()

	amqpConnection := amqp.NewAMQPConnection(&amqp.Config{
		User:     config.AMQPUser,
		Password: config.AMQPPassword,
		Host:     config.AMQPHost,
	}, logger)

	stopChan := make(chan struct{})
	listenForKillSignal(stopChan)

	transactionalClient := connector.TransactionalClient()

	integrationEventTransport := integrationevent.NewIntegrationEventTransport(
		integrationevent.NewIntegrationEventHandler(logger),
	)
	amqpConnection.AddChannel(integrationEventTransport)

	eventStore := mysql.NewEventStore(transactionalClient)

	storedEventSender := initStoredEventSender(
		transactionalClient,
		eventStore,
		integrationEventTransport,
		logger,
		time.Duration(config.StoredEventSenderDelay)*time.Second,
	)

	defer storedEventSender.Stop()

	contentServiceClient, err := initContentServiceClient(config)
	if err != nil {
		return err
	}

	container := infrastructure.NewDependencyContainer(
		connector.TransactionalClient(),
		logger,
		contentServiceClient,
		eventStore,
		storedEventSender.Increment,
	)

	err = amqpConnection.Start()
	if err != nil {
		return err
	}
	defer amqpConnection.Stop()

	serviceApi := transport.NewPlaylistServiceServer(container)
	serverHub := server.NewHub(stopChan)

	baseServer := grpc.NewServer(grpc.UnaryInterceptor(makeGRPCUnaryInterceptor(logger)))
	playlistservice.RegisterPlayListServiceServer(baseServer, serviceApi)

	serverHub.AddServer(server.NewGrpcServer(
		baseServer,
		server.GrpcServerConfig{ServeAddress: config.ServeGRPCAddress},
		logger),
	)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	var httpServer *http.Server

	serverHub.AddServer(&server.FuncServer{
		ServeImpl: func() error {
			grpcGatewayMux := runtime.NewServeMux()
			opts := []grpc.DialOption{grpc.WithInsecure()}
			err2 := playlistservice.RegisterPlayListServiceHandlerFromEndpoint(ctx, grpcGatewayMux, config.ServeGRPCAddress, opts)
			if err2 != nil {
				return err2
			}

			router := mux.NewRouter()
			router.PathPrefix("/api/").Handler(grpcGatewayMux)

			router.HandleFunc("/resilience/ready", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = io.WriteString(w, http.StatusText(http.StatusOK))
			}).Methods(http.MethodGet)

			httpServer = &http.Server{
				Handler:      transport.NewLoggingMiddleware(router, logger),
				Addr:         config.ServeRESTAddress,
				WriteTimeout: 15 * time.Second,
				ReadTimeout:  15 * time.Second,
			}

			logger.Info("REST server started")
			return httpServer.ListenAndServe()
		},
		StopImpl: func() error {
			cancel()
			return httpServer.Shutdown(context.Background())
		},
	})

	return serverHub.Run()
}

func initLogger() (log.MainLogger, error) {
	return jsonlog.NewLogger(&jsonlog.Config{AppName: appID}), nil
}

func listenForKillSignal(stopChan chan<- struct{}) {
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
		<-ch
		stopChan <- struct{}{}
	}()
}

func makeGRPCUnaryInterceptor(logger log.Logger) grpc.UnaryServerInterceptor {
	loggerInterceptor := transport.NewLoggerServerInterceptor(logger)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		resp, err = loggerInterceptor(ctx, req, info, handler)
		return resp, err
	}
}

func initContentServiceClient(config *config) (contentserviceapi.ContentServiceClient, error) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial(config.ContentServiceGRPCAddress, opts...)
	if err != nil {
		return nil, err
	}

	return contentserviceapi.NewContentServiceClient(conn), nil
}

func initStoredEventSender(
	client commonmysql.TransactionalClient,
	eventStore storedevent.Store,
	integrationEvenTransport storedevent.Transport,
	logger log.Logger,
	delay time.Duration,
) storedevent.Sender {
	tracker := mysql.NewEventsDispatchTracker(client)

	return storedevent.NewStoredEventSender(
		eventStore,
		tracker,
		integrationEvenTransport,
		delay,
		func(err error) { logger.Error(err) },
	)
}
