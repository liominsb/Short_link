package config

import (
	"Short_link/global"
	"log"

	"github.com/go-redis/redis"
)

func initRedis() {
	RedisCilnet := redis.NewClient(&redis.Options{
		Addr:         Appconf.Database.Addr,
		Password:     Appconf.Database.Password, // no password set
		DB:           0,                         // use default DB
		MinIdleConns: 500,                       //设置最小空闲连接数为3
		PoolSize:     500,                       //设置连接池大小为10
	})

	_, err := RedisCilnet.Ping().Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	global.RedisDB = RedisCilnet
}
