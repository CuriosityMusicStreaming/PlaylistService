package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/auth"
	log "github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/logger"
	jsonlog "github.com/CuriosityMusicStreaming/ComponentsPool/pkg/infrastructure/logger"
	commonserver "github.com/CuriosityMusicStreaming/ComponentsPool/pkg/infrastructure/server"
	"google.golang.org/grpc"

	"playlistservice/api/authorizationservice"
	contentserviceapi "playlistservice/api/contentservice"
	playlistserviceapi "playlistservice/api/playlistservice"
	"playlistservice/pkg/integrationtests/app"
	"playlistservice/pkg/integrationtests/infrastructure"
)

var appID = "UNKNOWN"

func main() {
	logger := initLogger()

	config, err := parseConfig()
	if err != nil {
		logger.FatalError(err)
	}

	err = runService(config, logger)
	if err != nil {
		logger.FatalError(err)
	}
}

func runService(config *config, logger log.MainLogger) error {
	server, userContainer := infrastructure.NewAuthorizationServer()

	baseServer := grpc.NewServer()
	authorizationservice.RegisterAuthorizationServiceServer(baseServer, server)

	grpcServer := commonserver.NewGrpcServer(baseServer,
		commonserver.GrpcServerConfig{ServeAddress: config.ServeGRPCAddress},
		logger,
	)

	runServer(grpcServer, logger)

	waitForService(config.PlaylistServiceHost + config.PlaylistServiceRESTAddress)
	waitForService(config.ContentServiceHost + config.ContentServiceRESTAddress)

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	playlistServiceClient, err := initPlaylistServiceClient(opts, config)
	if err != nil {
		return err
	}

	contentServiceClient, err := initContentServiceClient(opts, config)

	logger.Info("Start tests")

	userDescriptorSerializer := auth.NewUserDescriptorSerializer()

	app.RunTests(
		infrastructure.NewPlaylistServiceApi(
			playlistServiceClient,
			userDescriptorSerializer,
		),
		infrastructure.NewContentServiceApi(
			contentServiceClient,
			userDescriptorSerializer,
		),
		userContainer,
	)

	logger.Info("All test passed successfully")

	return nil
}

func waitForService(serviceAddress string) {
	const readyPath = "/resilience/ready"
	const retries = 30

	request, err := http.NewRequest(http.MethodGet, "http://"+serviceAddress+readyPath, nil)
	if err != nil {
		panic(err)
	}

	for i := 0; i < retries; i++ {
		res, err := http.DefaultClient.Do(request)
		if err == nil && res.StatusCode == http.StatusOK {
			_ = res.Body.Close()
			return
		}
		time.Sleep(time.Second)
	}
	panic("failed to wait service")
}

func initLogger() log.MainLogger {
	return jsonlog.NewLogger(&jsonlog.Config{AppName: appID})
}

func initPlaylistServiceClient(commonOpts []grpc.DialOption, config *config) (playlistserviceapi.PlayListServiceClient, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s%s", config.PlaylistServiceHost, config.PlaylistServiceGRPCAddress), commonOpts...)
	if err != nil {
		return nil, err
	}

	return playlistserviceapi.NewPlayListServiceClient(conn), nil
}

func initContentServiceClient(commonOpts []grpc.DialOption, config *config) (contentserviceapi.ContentServiceClient, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s%s", config.ContentServiceHost, config.PlaylistServiceGRPCAddress), commonOpts...)
	if err != nil {
		return nil, err
	}

	return contentserviceapi.NewContentServiceClient(conn), nil
}

func runServer(server commonserver.Server, logger log.MainLogger) {
	go func() {
		err := server.Serve()
		if err != nil {
			logger.FatalError(err, "failed to serve grpc")
		}
	}()
}
