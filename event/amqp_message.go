package event

import "github.com/streadway/amqp"

// AMQP消息事件
type AMQPMessageEvent struct {
	QueueName string
	D         *amqp.Delivery
}
