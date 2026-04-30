package global

import (
	"Short_link/utils"

	"github.com/go-redis/redis"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

var (
	Db          *gorm.DB //database 数据库
	RedisDB     *redis.Client
	GID         *utils.IDGenerator
	Conn        *amqp.Connection
	Queue       amqp.Queue
	ChannelPool chan *amqp.Channel
)
