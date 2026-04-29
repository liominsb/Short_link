package global

import (
	"Short_link/utils"

	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

var (
	Db      *gorm.DB //database 数据库
	RedisDB *redis.Client
	GID     *utils.IDGenerator
)
