package redis

import (
	"github.com/go-redis/redis"
	"github.com/zfiona/server-base/log"
)

var (
	rdb      *redis.Client
)

type Config struct {
	Addr string     //"127.0.0.1:6379"
	Password string //"123456"
	Block int       //6
}

func OpenDB(c *Config) {
	log.Debug("redis->open db")
	rdb = redis.NewClient(&redis.Options{
		Addr:     c.Addr,
		Password: c.Password,
		DB:       c.Block,  // use default DB
	})
}

func Db() *redis.Client {
	return rdb
}
