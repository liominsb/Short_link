package controllers

import (
	"Short_link/global"
	"Short_link/models"
	"Short_link/utils"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	amqp "github.com/rabbitmq/amqp091-go"
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
		err := global.Db.Model(&models.ShortLink{}).Where("id=?", utils.DecodeBase62(key)).First(&shortLink).Error
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

//func CreateRedirect(ctx *gin.Context) {
//	var input struct {
//		Url string `json:"url" form:"url" binding:"required"`
//	}
//
//	if err := ctx.ShouldBindJSON(&input); err != nil {
//		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//		return
//	}
//
//	key, err := global.GID.NextID()
//	if err != nil {
//		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//
//	keystr := utils.EncodeBase62(key)
//
//	err = global.Db.Create(&models.ShortLink{
//		ID:      key,
//		LongUrl: input.Url,
//	}).Error
//	if err != nil {
//		// 处理其他未知的数据库错误（不要把 err.Error() 抛给前端）
//		// 可以在后端打 log 记录真实的 err
//		log.Println("CreateRedirect,mysql", err)
//		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "系统内部错误"})
//		return
//	}
//
//	err = global.RedisDB.Set("key:"+keystr, input.Url, 24*time.Hour).Err()
//	if err != nil {
//		log.Println("CreateRedirect,Redis", err)
//	}
//
//	ctx.JSON(http.StatusOK, gin.H{"message": "短链接创建成功:" + keystr})
//}

// MsgPayload 定义要在 MQ 中传输的数据结构
type MsgPayload struct {
	ID      uint64 `json:"id"`
	LongURL string `json:"long_url"`
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

	err = global.RedisDB.Set("key:"+utils.EncodeBase62(key), input.Url, 1*time.Hour).Err()
	if err != nil {
		// Redis 写入失败时，为了保证数据一致性，通常选择直接报错，终止当前请求
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "缓存写入失败，请重试"})
		return
	}

	// 1. 组装要发给 MQ 的数据
	msg := MsgPayload{
		ID:      key,
		LongURL: input.Url,
	}
	body, _ := json.Marshal(msg)

	// 2. 将数据推送到 RabbitMQ
	ctx1, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = global.Channel.PublishWithContext(ctx1,
		"",                // exchange (默认交换机)
		global.Queue.Name, // routing key (直接发给我们的队列)
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent, // 消息持久化
			ContentType:  "application/json",
			Body:         body,
		})

	if err != nil {
		ctx.JSON(500, gin.H{"error": "系统繁忙，消息投递失败"})
		return
	}

	// 3. 立刻返回成功，绝不等待数据库
	ctx.JSON(200, gin.H{"short_url": utils.EncodeBase62(key)})
}
