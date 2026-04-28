package main

import (
	"Short_link/config"
	"Short_link/controllers"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	config.InitConfig()
	r := gin.Default()
	r.GET("/s/:key", controllers.Redirect)
	r.POST("/s", controllers.CreateRedirect)
	err := r.Run(config.Appconf.App.Port)
	if err != nil {
		log.Fatal(err)
	}
}
