package global

import (
	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

var (
	Db      *gorm.DB //database 数据库
	RedisDB *redis.Client
)
