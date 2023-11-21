package consume

import (
	"context"
	"strings"
	"sync"

	"example.com/demo/internal/daemon/config"
	"example.com/demo/internal/pkg/consumer"
	"example.com/demo/internal/pkg/consumer/process"
	"example.com/demo/pkg/log"
)

func Run(ctx context.Context, cfg *config.Config) {
	var wg sync.WaitGroup
	processCfg := process.NewProcessConfig(process.WithDefaultMQ(cfg.MQDefault), process.WithAppMode(cfg.AppMode), process.WithQueueRegion(cfg.QueueRegion))
	consumerGroups := strings.Split(cfg.ConsumerGroups, ",")

	for key, consumeConfig := range cfg.MQConsumer {
		if !filterConsumerGroup(consumerGroups, consumeConfig.Group) {
			continue
		}
		wg.Add(1)
		log.Infof("start consumer: %s %+v", key, consumeConfig)

		consumer := consumer.New(processCfg)
		consumer.Init(consumeConfig)
		go consumer.Start(ctx, &wg)
	}

	wg.Wait()
	log.Info("consumer done")
}

func filterConsumerGroup(consumerGroups []string, group string) bool {
	for _, v := range consumerGroups {
		if v == group {
			return true
		}
	}

	return false
}
