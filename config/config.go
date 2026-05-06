package config

import (
	"Short_link/MQ"
	"Short_link/global"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Port string
	}
	Database struct {
		Dsn          string
		MaxIdleConns int
		MaxOpenConns int
		Addr         string
		Password     string
		SubSwitch    bool //是否开启redis分布式锁，默认为false
	}
	TokenBucket struct {
		Rate     float64
		Capacity float64
	}
	Mq struct {
		Url      string
		PoolSize int
	}
}

var Appconf *Config

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	Appconf = &Config{}

	if err := viper.Unmarshal(Appconf); err != nil {
		log.Fatalf("Failed to unmarshal config file: %v", err)
	}

	initDB()
	initRedis()
	InitRabbitMQ()
	MQ.StartInsertWorker(global.Db)
}
