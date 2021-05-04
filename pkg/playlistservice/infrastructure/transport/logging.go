package transport

import (
	"context"
	"fmt"
	log "github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/logger"
	"google.golang.org/grpc"
	"net/http"
	"strings"
	"time"
)

func NewLoggingMiddleware(h http.Handler, logger log.Logger) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		now := time.Now()
		h.ServeHTTP(writer, request)

		logger.WithFields(log.Fields{
			"duration": time.Since(now),
			"method":   request.Method,
			"url":      request.RequestURI,
		}).Info("request finished")
	})
}

func NewLoggerServerInterceptor(logger log.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()

		resp, err = handler(ctx, req)

		fields := log.Fields{
			"args":     req,
			"duration": fmt.Sprintf("%v", time.Since(start)),
			"method":   getGRPCMethodName(info),
		}

		entry := logger.WithFields(fields)
		if err != nil {
			entry.Error(err, "call failed")
		} else {
			entry.Info("call finished")
		}

		return resp, translateError(err)
	}
}

func getGRPCMethodName(info *grpc.UnaryServerInfo) string {
	method := info.FullMethod
	return method[strings.LastIndex(method, "/")+1:]
}
