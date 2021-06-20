## Playlist service

Service to organize songs and podcasts to playlists

### Dependencies

This service depend on so-called [Platform](https://github.com/CuriosityMusicStreaming/Platform). 
It provides local environment and necessary devtools(like [apisynchronizer](https://github.com/UsingCoding/ApiSynchronizer) to sync api files between services)

Microservices that depending on:
* [ContentService](https://github.com/CuriosityMusicStreaming/ContentService)
  
Other libraries:
* [ComponentsPool](https://github.com/CuriosityMusicStreaming/ComponentsPool) - common library with components
* [ApiStore](https://github.com/CuriosityMusicStreaming/ApiStore) - repository that provides services api that synced by apisynchronizer
* [Protobuf](https://github.com/protocolbuffers/protobuf) - provides protobuf api codegen
* [GrpcGateway](https://github.com/grpc-ecosystem/grpc-gateway) - v1 only - provides rest proxy to grpc server
* Other code dependencies in `go.mod` 

### Build

**To have ability to build service download [Platform](https://github.com/CuriosityMusicStreaming/Platform) and make installation steps**

Run make

```shell
make
```

Command build all dependencies and put binary file to `bin/` 

Run `make publish` to dockerize service


### Test

You can run unit-tests
```shell
make test
```

You can run linter
```shell
make check
```

You can run integration-tests

```shell
make build publish && ./bin/run-integraion-tests.sh
```

#### Integration-tests

Tests that checks integration with other services(ContentService and MySQL) and full user-cases with service