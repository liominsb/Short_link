package MQ

import (
	"Short_link/global"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func InitRabbitMQ() {
	var err error
	// 1. 建立连接
	global.Conn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("连接 RabbitMQ 失败: %v", err)
	}

	// 2. 打开通道
	global.Channel, err = global.Conn.Channel()
	if err != nil {
		log.Fatalf("打开 Channel 失败: %v", err)
	}

	// 3. 声明队列（持久化开启）
	global.Queue, err = global.Channel.QueueDeclare(
		"short_link_insert_queue", // 队列名称
		true,                      // durable: 队列持久化
		false,                     // delete when unused
		false,                     // exclusive
		false,                     // no-wait
		nil,                       // arguments
	)
	if err != nil {
		log.Fatalf("声明队列失败: %v", err)
	}
}
