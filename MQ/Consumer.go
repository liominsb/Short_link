package MQ

import (
	"Short_link/controllers"
	"Short_link/global"
	"Short_link/models"
	"encoding/json"
	"log"
	"time"

	"gorm.io/gorm"
)

// StartInsertWorker 启动异步写库协程
func StartInsertWorker(db *gorm.DB) {
	ch, err := global.Conn.Channel()
	if err != nil {
		log.Fatalf("create consumer channel failed: %v", err)
	}

	msgs, err := ch.Consume(
		global.Queue.Name,
		"worker_1",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		log.Fatalf("注册消费者失败: %v", err)
	}

	go func() {
		defer func() {
			if err := ch.Close(); err != nil {
				log.Printf("close consumer channel failed: %v", err)
			}
		}()

		var batch []models.ShortLink // 用于暂存数据的切片
		var deliveryTags []uint64    // 记录这些消息的 Tag，用于最后确认

		batchSize := 1000                         // 攒够 1000 条写一次数据库
		ticker := time.NewTicker(1 * time.Second) // 或者最多等 1 秒写一次数据库
		defer ticker.Stop()

		flush := func() {
			if len(batch) == 0 {
				return
			}

			if err := db.CreateInBatches(batch, len(batch)).Error; err != nil {
				log.Printf("批量落库失败: %v", err)
				return
			}

			if err := ch.Ack(deliveryTags[len(deliveryTags)-1], true); err != nil {
				log.Printf("ack messages failed: %v", err)
				return
			}

			batch = batch[:0]
			deliveryTags = deliveryTags[:0]
		}

		for {
			select {
			case d, ok := <-msgs:
				if !ok {
					return
				}

				var msg controllers.MsgPayload // 解析 JSON
				if err := json.Unmarshal(d.Body, &msg); err == nil {
					batch = append(batch, models.ShortLink{
						ID:      msg.ID,
						LongUrl: msg.LongURL,
					})
					deliveryTags = append(deliveryTags, d.DeliveryTag)
				}

				if len(batch) >= batchSize {
					flush()
				}
			case <-ticker.C:
				flush()
			}
		}
	}()
}
