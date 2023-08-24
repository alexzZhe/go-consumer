package mq

import (
	"context"
	"sync"

	"example.com/demo/pkg/log"
	"github.com/ThreeDotsLabs/watermill-amqp/v2/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	uuid "github.com/satori/go.uuid"
)

type RabbitMQConfig struct {
	AmqpURI         string `json:"amqp_uri"       mapstructure:"amqp_uri"`
	ChannelPoolSize int    `json:"channel_pool_size"  mapstructure:"channel_pool_size"`
	ConfirmDelivery bool   `json:"confirm_delivery"   mapstructure:"confirm_delivery"`
	ExchangeName    string `json:"exchange_name"      mapstructure:"exchange_name"`
}

type RabbitMQ struct {
	// config     *RabbitMQConfig
	conn       *amqp.ConnectionWrapper
	amqpConfig amqp.Config
}

var (
	oncePublisher sync.Once
	publisher     *amqp.Publisher
)

func NewRabbitMQConfig() *RabbitMQConfig {
	return &RabbitMQConfig{
		AmqpURI:         "amqp://guest:guest@localhost:5672/",
		ChannelPoolSize: 1,
		ConfirmDelivery: true,
		ExchangeName:    "",
	}
}

func newDurableRoutingConfig(cfg *RabbitMQConfig) amqp.Config {
	return amqp.Config{
		Connection: amqp.ConnectionConfig{
			AmqpURI: cfg.AmqpURI,
		},

		Marshaler: amqp.DefaultMarshaler{},

		Exchange: amqp.ExchangeConfig{
			GenerateName: func(topic string) string {
				return cfg.ExchangeName
			},
			Type:    "direct",
			Durable: true,
		},
		Queue: amqp.QueueConfig{
			GenerateName: func(topic string) string {
				return topic
			},
			Durable: true,
		},
		QueueBind: amqp.QueueBindConfig{
			GenerateRoutingKey: func(topic string) string {
				return topic
			},
		},
		Publish: amqp.PublishConfig{
			GenerateRoutingKey: func(topic string) string {
				return topic
			},
			ChannelPoolSize: cfg.ChannelPoolSize,
			ConfirmDelivery: cfg.ConfirmDelivery,
		},
		Consume: amqp.ConsumeConfig{
			Qos: amqp.QosConfig{
				PrefetchCount: 1,
			},
		},
		TopologyBuilder: &amqp.DefaultTopologyBuilder{},
	}
}

func NewRabbitMQ(cfg *RabbitMQConfig) MQ {
	amqpConfig := newDurableRoutingConfig(cfg)
	conn, err := amqp.NewConnection(amqpConfig.Connection, newMQLogger())
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	return &RabbitMQ{conn: conn, amqpConfig: amqpConfig}
}

func (mq *RabbitMQ) publisher() (*amqp.Publisher, error) {
	var err error
	oncePublisher.Do(func() {
		publisher, err = amqp.NewPublisherWithConnection(mq.amqpConfig, newMQLogger(), mq.conn)
		if err != nil {
			log.Fatalf("Failed to create publisher: %v", err)
		}
	})

	return publisher, err
}

func (mq *RabbitMQ) SendMessage(ctx context.Context, queueName string, messages ...*Message) error {
	pub, err := mq.publisher()
	uuidStr := uuid.Must(uuid.NewV4(), nil).String()

	for _, msg := range messages {
		m := message.NewMessage(uuidStr, []byte(msg.Content))
		err = pub.Publish(queueName, m)
		if err != nil {
			log.Errorf("Failed to publish message: %v", err)
			return err
		}
	}

	return err
}

func (mq *RabbitMQ) ConsumeMessage(ctx context.Context, c Consumer, callback func(interface{}) bool) error {
	sub, err := amqp.NewSubscriberWithConnection(mq.amqpConfig, newMQLogger(), mq.conn)
	if err != nil {
		log.Fatalf("Failed to create subscriber: %v", err)
	}
	messages, err := sub.Subscribe(ctx, c.QueueName)
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	for msg := range messages {
		// log.Infof("received message: %s, payload: %s\n", msg.UUID, string(msg.Payload))
		if callback([]byte(msg.Payload)) {
			msg.Ack()
		}

	}

	return err
}

func (mq *RabbitMQ) Close() error {

	// if publisher != nil {
	// 	publisher.Close()
	// }

	log.Infof("Closing RabbitMQ connection")
	mq.conn.Close()
	log.Infof("RabbitMQ connection closed")

	return nil
}
