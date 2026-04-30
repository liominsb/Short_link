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

//无MQ版本
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

	// 【探针 1：测 MySQL 发号器步长性能】
	t1 := time.Now()
	key, err := global.GID.NextID()
	log.Printf("探针1 - 生成ID耗时: %v\n", time.Since(t1))

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 【探针 2：测 Redis 网络或写入性能】
	t2 := time.Now()
	err = global.RedisDB.Set("key:"+utils.EncodeBase62(key), input.Url, 1*time.Hour).Err()
	log.Printf("探针2 - Redis写入耗时: %v\n", time.Since(t2))

	if err != nil {
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

	ch := <-global.ChannelPool
	defer func() { global.ChannelPool <- ch }()

	// 【探针 3：测 RabbitMQ 磁盘刷盘或网络投递性能】
	t3 := time.Now()
	err = ch.PublishWithContext(ctx1,
		"",                // exchange
		global.Queue.Name, // routing key
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent, // 消息持久化
			ContentType:  "application/json",
			Body:         body,
		})
	log.Printf("探针3 - MQ投递耗时: %v\n", time.Since(t3))

	if err != nil {
		log.Printf("MQ投递失败, ID: %d, Error: %v\n", key, err)
		ctx.JSON(500, gin.H{"error": "系统繁忙，消息投递失败"})
		return
	}

	// 3. 立刻返回成功
	ctx.JSON(200, gin.H{"short_url": utils.EncodeBase62(key)})
}
