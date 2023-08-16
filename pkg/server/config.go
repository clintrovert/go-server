package server

import (
	"time"

	"github.com/clintrovert/go-server/pkg/server/interceptors"
)

type CacheConfig struct {
	Cache      interceptors.KeyValCache
	KeyGenFunc interceptors.KeyGenerationFunc
	Ttl        time.Duration
}
