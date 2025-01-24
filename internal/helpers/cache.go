package helpers

import (
	"fmt"

	"github.com/arizon-dread/sse/internal/config"
	"github.com/redis/go-redis/v9"
)

func GetCacheConn() (*redis.Client, error) {
	conf := config.Get()
	if conf.Cache.Url == "" {
		return nil, fmt.Errorf("no url to cache supplied in settings")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.Cache.Url,
		Password: conf.Cache.Password,
	})

	return rdb, nil
}
