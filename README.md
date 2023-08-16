# go-server

```golang
func main() {
   ...
   srv, err := server.NewBuilder(grpcPort, log).
         WithRateLimiter(limiter).
         WithMetrics(httpPort, prometheus.DefaultRegister).
         WithResponseCache(cache).
         WithGrpcReflection().
         Build()
	
   if err != nil {
      ...
   }

   // Can access http server
   srv.HttpServer.ReadHeaderTimeout = time.Second * 2
	
   // Can access grpc server
   example.RegisterExampleService(srv.GrpcServer)
   ...
   srv.Serve()
}
```