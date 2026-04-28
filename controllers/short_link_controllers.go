package controllers

import (
	"Short_link/global"
	"Short_link/models"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

func Redirect(ctx *gin.Context) {
	key := ctx.Param("key")
	url, err := global.RedisDB.Get("key:" + key).Result()
	if err != nil {
		log.Println("Redirect,Redis", err)
	}

	shortLink := models.ShortLink{}

	if errors.Is(err, redis.Nil) {
		err := global.Db.Model(&models.ShortLink{}).Where("short_key=?", key).First(&shortLink).Error
		if err != nil {
			global.RedisDB.Set("key:"+key, shortLink.LongUrl, 60*time.Second)
			ctx.JSON(http.StatusNotFound, gin.H{"error": "短链接不存在"})
			return
		}
		err = global.RedisDB.Set("key:"+key, shortLink.LongUrl, 24*time.Hour).Err()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		url = shortLink.LongUrl
	}

	ctx.Redirect(http.StatusFound, url)
}

func CreateRedirect(ctx *gin.Context) {
	var input struct {
		Key string `json:"key" form:"key" binding:"required"`
		Url string `json:"url" form:"url" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := global.Db.Create(&models.ShortLink{
		ShortKey: input.Key,
		LongUrl:  input.Url,
	}).Error
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = global.RedisDB.Set("key:"+input.Key, input.Url, 24*time.Hour).Err()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "短链接创建成功"})
}
