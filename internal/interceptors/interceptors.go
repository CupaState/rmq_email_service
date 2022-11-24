package interceptors

import (
	"context"
	"net/http"
	"rmq_service/config"
	"rmq_service/pkg/grpc_errors"
	"rmq_service/pkg/logger"
	"rmq_service/pkg/metrics"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// InterceptorManager
type InteceptorManager struct {
	logger 	logger.Logger
	cfg 		*config.Config
	metr 		metrics.Metrics
}

// InterceptorManager Constructor
func NewInterceptorManager(logger logger.Logger, cfg *config.Config, metr metrics.Metrics) *InteceptorManager {
	return &InteceptorManager{logger: logger, cfg: cfg, metr: metr}
}

// Logger Interceptor
func (im *InteceptorManager) Logger(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	start := time.Now()
	md, _ := metadata.FromIncomingContext(ctx)
	reply, err := handler(ctx, req)
	im.logger.Infof("Method: %s, Time: %v, Metadata: %v, Err: %v",
		info.FullMethod,
		time.Since(start),
		md,
		err,
	)

	return reply, err
}

func (im *InteceptorManager) Metrics(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	var status = http.StatusOK

	if err != nil {
		status = grpc_errors.MapGRPCErrCodeToHttpStatus(grpc_errors.ParseGRPCErrStatusCode(err))
	}

	im.metr.ObserveResponseTime(status, info.FullMethod, info.FullMethod, time.Since(start).Seconds())
	im.metr.IncHits(status, info.FullMethod, info.FullMethod)
	return resp, err
}
