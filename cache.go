package main

import (
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/metrics"
	"github.com/eko/gocache/lib/v4/store"
	ristretto_store "github.com/eko/gocache/store/ristretto/v4"
	rueidis_store "github.com/eko/gocache/store/rueidis/v4"
	"github.com/redis/rueidis"
	"github.com/spf13/viper"

	"github.com/wei840222/ory-oathkeeper-login/config"
)

func NewCache() (cache.CacheInterface[string], error) {
	var s store.StoreInterface

	if viper.GetString(config.KeyCacheRedisHost) != "" {
		rueidisClient, err := rueidis.NewClient(rueidis.ClientOption{
			InitAddress: []string{fmt.Sprintf("%s:%d", viper.GetString(config.KeyCacheRedisHost), viper.GetInt(config.KeyCacheRedisPort))},
			Password:    viper.GetString(config.KeyCacheRedisPassword),
			SelectDB:    viper.GetInt(config.KeyCacheRedisDB),
		})
		if err != nil {
			return nil, err
		}

		s = rueidis_store.NewRueidis(rueidisClient, store.WithClientSideCaching(15*time.Second))
	} else {
		ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
			NumCounters: 1000,
			MaxCost:     100,
			BufferItems: 64,
		})
		if err != nil {
			return nil, err
		}

		s = ristretto_store.NewRistretto(ristrettoCache)
	}

	p := metrics.NewPrometheus(config.AppName)
	c := cache.New[string](s)

	return cache.NewMetric(p, c), nil
}
