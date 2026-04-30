package main

import (
	"Short_link/config"
	"Short_link/controllers"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	config.InitConfig()
	gin.SetMode(gin.ReleaseMode)

	//弃用 gin.Default()，改用 gin.New() 创建一个没有任何中间件的纯净引擎
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/s/:key", controllers.Redirect)
	r.POST("/s", controllers.CreateRedirect)
	err := r.Run(config.Appconf.App.Port)
	if err != nil {
		log.Fatal(err)
	}
}
