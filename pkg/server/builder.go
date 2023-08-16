package server

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/clintrovert/go-server/pkg/server/interceptors"
	openmetrics "github.com/grpc-ecosystem/go-grpc-middleware/providers/openmetrics/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/ratelimit"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	tcpPortMin = 1
	tcpPortMax = 65535
)

var (
	ErrCouldNotBuild = errors.New("builder could not be built")

	errGrpcPortAboveMax = errors.New("grpc port above max port number")
	errGrpcPortBelowMin = errors.New("grpc port below min port number")
	errHttpPortAboveMax = errors.New("http port above max port number")
	errHttpPortBelowMin = errors.New("http port below min port number")
	errLogEntryNotFound = errors.New("log entry was not assigned")
	errLimiterNotFound  = errors.New("rate limiter was not assigned")
)

type Builder struct {
	errs               []error
	grpcPort, httpPort int
	reflectionEnabled  bool
	metricsEnabled     bool
	log                *logrus.Entry

	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
}

// NewBuilder provides construction of a gRPC server with optional
// HTTP metrics server via chaining interceptor declarations.
func NewBuilder(grpcPort int, log *logrus.Entry) *Builder {
	b := &Builder{}

	if grpcPort > tcpPortMax {
		b.errs = append(b.errs, errGrpcPortAboveMax)
	}
	if grpcPort < tcpPortMin {
		b.errs = append(b.errs, errGrpcPortBelowMin)
	}
	b.grpcPort = grpcPort

	if log == nil {
		b.errs = append(b.errs, errLogEntryNotFound)
	}
	b.log = log

	return b
}

// WithRateLimiter todo: describe
func (b *Builder) WithRateLimiter(limiter ratelimit.Limiter) *Builder {
	if len(b.errs) > 0 {
		// Exit early if errors are already found.
		return b
	}

	if limiter == nil {
		b.errs = append(b.errs, errLimiterNotFound)
		return b
	}

	b.unaryInterceptors = append(
		b.unaryInterceptors,
		ratelimit.UnaryServerInterceptor(limiter),
	)

	b.streamInterceptors = append(
		b.streamInterceptors,
		ratelimit.StreamServerInterceptor(limiter),
	)

	return b
}

// WithResponseCache todo: describe
func (b *Builder) WithResponseCache(conf CacheConfig) *Builder {
	if len(b.errs) > 0 {
		// Exit early if errors are already found.
		return b
	}

	i, err := interceptors.NewCacheInterceptor(conf.Cache, b.log)
	if err != nil {
		b.errs = append(b.errs, err)
		return b
	}

	b.unaryInterceptors = append(
		b.unaryInterceptors,
		i.UnaryServerInterceptor(conf.KeyGenFunc, conf.Ttl),
	)

	b.streamInterceptors = append(
		b.streamInterceptors,
		i.StreamServerInterceptor(conf.KeyGenFunc, conf.Ttl),
	)

	return b
}

// WithMetrics todo: describe
func (b *Builder) WithMetrics(
	httpPort int,
	registerer prometheus.Registerer,
) *Builder {
	if len(b.errs) > 0 {
		// Exit early if errors are already found.
		return b
	}

	if httpPort > tcpPortMax {
		b.errs = append(b.errs, errHttpPortAboveMax)
		return b
	}

	if httpPort < tcpPortMin {
		b.errs = append(b.errs, errHttpPortBelowMin)
		return b
	}

	srvMetrics := openmetrics.NewRegisteredServerMetrics(
		registerer,
		openmetrics.WithServerHandlingTimeHistogram(),
	)
	registry := prometheus.NewRegistry()
	registry.MustRegister(srvMetrics)

	b.log.
		WithField("http_port", httpPort).
		Traceln("registering metrics interceptor")

	b.unaryInterceptors = append(
		b.unaryInterceptors,
		openmetrics.UnaryServerInterceptor(srvMetrics),
	)

	b.streamInterceptors = append(
		b.streamInterceptors,
		openmetrics.StreamServerInterceptor(srvMetrics),
	)
	b.httpPort = httpPort

	return b
}

// WithGrpcReflection todo: describe
func (b *Builder) WithGrpcReflection() *Builder {
	b.reflectionEnabled = true

	return b
}

// Build todo: describe
func (b *Builder) Build() (*Server, error) {
	if len(b.errs) > 0 {
		for _, err := range b.errs {
			b.log.
				WithError(err).
				Error("builder.Build(), error occurred")
		}

		return nil, ErrCouldNotBuild
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(b.unaryInterceptors...),
		grpc.ChainStreamInterceptor(b.streamInterceptors...),
	)

	if b.reflectionEnabled {
		reflection.Register(grpcServer)
	}

	var httpServer *http.Server
	if b.metricsEnabled {
		httpServer = &http.Server{
			Addr: strconv.Itoa(b.httpPort),
		}
	}

	return &Server{
		GrpcServer: grpcServer,
		HttpServer: httpServer,
	}, nil
}
