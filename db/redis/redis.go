package redis

import (
	"github.com/go-redis/redis"
	"github.com/zfiona/server-base/log"
)

var (
	rdb      *redis.Client
)

func OpenDB() {
	log.Debug("redis->open db")
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost",
		Password: "123456",
		DB:       6,  // use default DB
	})
}

func Db() *redis.Client {
	return rdb
}
