package config

import (
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
}
