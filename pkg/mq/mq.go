package mq

import "context"

const (
	AliyunMNSServer = "AliyunMNS"
	KafkaServer     = "Kafka"
	RabbitMQServer  = "RabbitMQ"
	RedisServer     = "Redis"
)

type Message struct {
	Key          string
	Content      string
	DelaySeconds int
	Priority     int
}

type Consumer struct {
	QueueName string
	GroupID   string
	AutoAck   bool
}

type Client interface {
	Close() error
}

type MQ interface {
	SendMessage(ctx context.Context, queueName string, messages ...*Message) error
	// ReceiveMessage() error
	ConsumeMessage(ctx context.Context, c Consumer, callback func(interface{}) bool) error
	Client
}
