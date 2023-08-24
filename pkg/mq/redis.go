package mq

import (
	"context"
	"errors"
	"sync"
	"time"

	redis "github.com/go-redis/redis/v7"

	"example.com/demo/pkg/log"
	"example.com/demo/pkg/storage"
)

var (
	connectOnce  sync.Once
	consumeRedis sync.WaitGroup
)

type RedisMQ struct {
	client *storage.RedisCluster
}

func NewRedisMQ(config *storage.Config) MQ {
	client := &storage.RedisCluster{}

	connectOnce.Do(func() {
		connectRedis(config)
	})

	return &RedisMQ{client: client}
}

func connectRedis(config *storage.Config) {

	var connecting bool

	for {
		if storage.Connected() {
			log.Info("Redis is already connected")
			break
		} else if !connecting {
			// ctx, cancel := context.WithCancel(context.Background())
			connecting = true
			go storage.ConnectToRedis(context.Background(), config)
		}

		isConnected := <-storage.ConnectStatus()
		if !isConnected {
			log.Errorf("Failed to connect to redis")
			break
		} else {
			log.Infof("Connected to redis")
		}
	}
}

func (mq *RedisMQ) SendMessage(ctx context.Context, queueName string, messages ...*Message) error {
	var message []string
	for _, msg := range messages {
		message = append(message, msg.Content)

	}
	err := mq.client.LPush(queueName, message...)
	if err != nil {
		log.Errorf("Failed to send message: %v", err)
	}

	return err
}

func (mq *RedisMQ) ConsumeMessage(ctx context.Context, c Consumer, callback func(interface{}) bool) error {
	consumeRedis.Add(1)
	defer consumeRedis.Done()

	for {
		select {
		case <-ctx.Done():
			log.Infof("Context canceled")
			return nil
		default:
			msg, err := mq.client.BRPop(c.QueueName, 10)

			if err != nil {
				if !errors.Is(err, redis.Nil) {
					log.Errorf("Failed to start BRPop: %+v", err)
				}
				if errors.Is(err, storage.ErrRedisIsDown) {
					time.Sleep(1 * time.Second)
				}

			} else {
				// log.Infof("Received message: %v", msg)
				if len(msg) > 1 {
					callback([]byte(msg[1]))
					// log.Infof("Received message: %v", msg[1])
				}
			}
		}
	}

	// return nil
}

func (mq *RedisMQ) Close() error {
	consumeRedis.Wait()
	return nil
}
