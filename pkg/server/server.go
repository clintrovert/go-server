package server

import (
	"net"
	"net/http"
	"os"

	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	tcp     = "tcp"
	metrics = "/metrics"
)

type Server struct {
	log                *logrus.Entry
	metricsEnabled     bool
	grpcPort, httpPort string
	GrpcServer         *grpc.Server
	HttpServer         *http.Server
	MetricsRegistry    *prometheus.Registry
}

func (srv *Server) Serve() {
	g := &run.Group{}
	g.Add(srv.serveGrpc())

	if srv.metricsEnabled {
		g.Add(srv.serveHttp())
	}

	if err := g.Run(); err != nil {
		os.Exit(1)
	}
}

func (srv *Server) serveGrpc() (func() error, func(error)) {
	return func() error {
			l, err := net.Listen(tcp, srv.grpcPort)
			if err != nil {
				return err
			}
			return srv.GrpcServer.Serve(l)
		}, func(err error) {
			srv.GrpcServer.GracefulStop()
			srv.GrpcServer.Stop()
		}
}

func (srv *Server) serveHttp() (func() error, func(error)) {
	return func() error {
			m := http.NewServeMux()
			if srv.MetricsRegistry != nil {
				m.Handle(metrics, promhttp.HandlerFor(
					srv.MetricsRegistry,
					promhttp.HandlerOpts{
						EnableOpenMetrics: true,
					},
				))
			}

			srv.HttpServer.Handler = m
			srv.log.
				WithField("http_port", srv.HttpServer.Addr).
				Infoln("http metrics server started")

			return srv.HttpServer.ListenAndServe()
		}, func(error) {
			if err := srv.HttpServer.Close(); err != nil {
				srv.log.
					WithError(err).
					Fatalln("failed to close http metrics server")
			}
		}
}
