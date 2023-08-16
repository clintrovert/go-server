package interceptors

import (
	"testing"
)

type cacheTestClient struct {
	interceptor CacheInterceptor
}

func newCacheTestClient(t *testing.T) *cacheTestClient {
	return &cacheTestClient{}
}
