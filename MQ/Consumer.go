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

	ch := <-global.ChannelPool
	defer func() { global.ChannelPool <- ch }()

	msgs, err := ch.Consume(
		global.Queue.Name,
		"worker_1", // 消费者名字
		false,      // auto-ack 设为 false，必须手动确认，防止程序崩溃丢数据
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("注册消费者失败: %v", err)
	}

	go func() {
		var batch []models.ShortLink // 用于暂存数据的切片
		var deliveryTags []uint64    // 记录这些消息的 Tag，用于最后确认

		batchSize := 1000                         // 攒够 1000 条写一次数据库
		ticker := time.NewTicker(1 * time.Second) // 或者最多等 1 秒写一次数据库
		defer ticker.Stop()

		// 抽离出一个执行落库的局部函数
		flush := func() {
			if len(batch) == 0 {
				return
			}
			// 批量插入
			if err := db.CreateInBatches(batch, len(batch)).Error; err != nil {
				log.Printf("批量落库失败: %v", err)
				return // 生产环境中这里需要做错误重试
			}
			// 告诉 MQ：这批数据我处理完了，可以从队列里删了
			ch.Ack(deliveryTags[len(deliveryTags)-1], true)

			// 清空切片，迎接下一批数据
			batch = batch[:0]
			deliveryTags = deliveryTags[:0]
		}

		// 死循环监听
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
