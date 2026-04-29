package controllers

import (
	"Short_link/global"
	"Short_link/models"
	"Short_link/utils"
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

	if err == nil && url != "" {
		ctx.Redirect(http.StatusFound, url)
		return
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
		Url string `json:"url" form:"url" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key, err := global.GID.NextID()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	keystr := utils.EncodeBase62(uint64(key))

	err = global.Db.Create(&models.ShortLink{
		ShortKey: keystr,
		LongUrl:  input.Url,
	}).Error
	if err != nil {
		// 处理其他未知的数据库错误（不要把 err.Error() 抛给前端）
		// 可以在后端打 log 记录真实的 err
		log.Println("CreateRedirect,mysql", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "系统内部错误"})
		return
	}

	err = global.RedisDB.Set("key:"+keystr, input.Url, 24*time.Hour).Err()
	if err != nil {
		log.Println("CreateRedirect,Redis", err)
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "短链接创建成功:" + keystr})
}
