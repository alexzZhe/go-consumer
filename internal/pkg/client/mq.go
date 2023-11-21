package client

import (
	"context"
	"fmt"
	"sync"

	genericoptions "example.com/demo/internal/pkg/options"
	"example.com/demo/pkg/log"
	"example.com/demo/pkg/mq"
	"example.com/demo/pkg/storage"
)

const (
	// DefaultMQServer is the default message queue server.
	DefaultMQ = "default"
)

type MQClient struct {
	redisConfig     *storage.Config
	aliyunMNSConfig *mq.MNSConfig
	kafkaConfig     *mq.KafkaConf
	rabbitMQConfig  *mq.RabbitMQConfig
	defaultMQServer string
	AppMode         string
	QueueRegion     string
}

type MQOptions func(*MQClient)

var (
	_ mq.MQ = (*mq.AliyunMNS)(nil)
	_ mq.MQ = (*mq.KafkaMQ)(nil)
	_ mq.MQ = (*mq.RabbitMQ)(nil)
	_ mq.MQ = (*mq.RedisMQ)(nil)

	MQFactory      = &MQClient{}
	redisMQClient  mq.MQ
	rabbitMQClient mq.MQ
	redisMQOnce    sync.Once
	rabbitMQOnce   sync.Once
)

func CreateMQ(server string) mq.MQ {
	c := MQFactory
	if server == DefaultMQ {
		server = c.defaultMQServer
	}

	switch server {
	case mq.AliyunMNSServer:
		return mq.NewAliyunMNS(c.aliyunMNSConfig)
	case mq.KafkaServer:
		return mq.NewKafkaMQ(c.kafkaConfig)
	case mq.RabbitMQServer:
		// return mq.NewRabbitMQ(c.rabbitMQConfig)
		// 单例
		rabbitMQOnce.Do(func() {
			rabbitMQClient = mq.NewRabbitMQ(c.rabbitMQConfig)
		})
		return rabbitMQClient
	case mq.RedisServer:
		// 单例
		redisMQOnce.Do(func() {
			redisMQClient = mq.NewRedisMQ(c.redisConfig)
		})
		return redisMQClient
	default:
		log.Errorf("unknown mq server: %s", server)
	}

	return nil
}

func SendToQueue(queueName string, messages ...*mq.Message) error {

	c := CreateMQ(DefaultMQ)

	err := c.SendMessage(context.TODO(), queueNameFormat(queueName), messages...)
	if err != nil {
		log.Errorf("Failed to send message: %v", err)
	}

	return err
}

func queueNameFormat(name string) string {
	if MQFactory.AppMode != "" {
		return fmt.Sprintf("%s-%s%s", name, MQFactory.AppMode, MQFactory.QueueRegion)
	}

	return name
}

func InitMQOptions(opts ...MQOptions) {
	c := MQFactory
	for _, opt := range opts {
		opt(c)
	}
	// return c
}

func WithDefaultMQ(server string) MQOptions {
	return func(c *MQClient) {
		c.defaultMQServer = server
	}
}

func WithAppMode(appMode string) MQOptions {
	return func(c *MQClient) {
		c.AppMode = appMode
	}
}

func WithQueueRegion(queueRegion string) MQOptions {
	return func(c *MQClient) {
		c.QueueRegion = queueRegion
	}
}

func WithRedisConfig(config *genericoptions.RedisOptions) MQOptions {
	cfg := &storage.Config{
		Host:                  config.Host,
		Port:                  config.Port,
		Addrs:                 config.Addrs,
		MasterName:            config.MasterName,
		Username:              config.Username,
		Password:              config.Password,
		Database:              config.Database,
		MaxIdle:               config.MaxIdle,
		MaxActive:             config.MaxActive,
		Timeout:               config.Timeout,
		EnableCluster:         config.EnableCluster,
		UseSSL:                config.UseSSL,
		SSLInsecureSkipVerify: config.SSLInsecureSkipVerify,
	}
	return func(c *MQClient) {
		c.redisConfig = cfg
	}
}

func WithAliyunMNSConfig(config *mq.MNSConfig) MQOptions {
	return func(c *MQClient) {
		c.aliyunMNSConfig = config
	}
}

func WithKafkaConfig(config *mq.KafkaConf) MQOptions {
	return func(c *MQClient) {
		c.kafkaConfig = config
	}
}

func WithRabbitMQConfig(config *mq.RabbitMQConfig) MQOptions {
	return func(c *MQClient) {
		c.rabbitMQConfig = config
	}
}
