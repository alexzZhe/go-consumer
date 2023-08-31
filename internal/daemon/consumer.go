package daemon

import (
	"context"
	"fmt"
	"sync"
	"time"

	"example.com/demo/internal/daemon/config"
	"example.com/demo/internal/daemon/options"
	"example.com/demo/internal/daemon/process"
	"example.com/demo/internal/pkg/client"
	"example.com/demo/pkg/log"
	"example.com/demo/pkg/mq"
	"github.com/panjf2000/ants/v2"
)

type ProcessConfig struct {
	Host string
}

type Consumer struct {
	ctx            context.Context
	config         *config.Config
	processCfg     *ProcessConfig
	processPool    *ants.Pool
	consumerConfig options.MQConsumerConfig
}

const (
	defaultMQ = "default"
)

func NewProcessConfig(cfg *config.Config) *ProcessConfig {
	processCfg := &ProcessConfig{
		Host: cfg.MessageProcessHost,
	}

	return processCfg
}

func NewConsumer(ctx context.Context, cfg *config.Config) *Consumer {
	processCfg := NewProcessConfig(cfg)
	return &Consumer{ctx: ctx, config: cfg, processCfg: processCfg}
}

func (c *Consumer) Init(opt options.MQConsumerConfig) {
	c.consumerConfig = opt

	// 按配置初始化工作池大小
	processPool, err := ants.NewPool(opt.MessageProcessorNum)
	c.processPool = processPool
	if err != nil {
		log.Errorf("Failed to create ants pool: %v", err)
	}
}

// 消费队列
func (c *Consumer) Start(wg *sync.WaitGroup) {
	defer wg.Done()

	queueName := c.queueName(c.consumerConfig.QueueName)
	if c.consumerConfig.ConsumerNum <= 0 {
		log.Errorf("%s ConsumerNum must be greater than 0", queueName)
		return
	}

	consumer := mq.Consumer{QueueName: queueName, GroupID: c.consumerConfig.GroupID}
	mqClients := c.newMQClient()
	processors := c.newProcessor()

	log.Infof("%s start consumerNum: %d messageProcessorNum: %d", queueName, c.consumerConfig.ConsumerNum, c.processPool.Cap())

	for i := 0; i < c.consumerConfig.ConsumerNum; i++ {
		go func(index int) {
			log.Infof("%d consume goroutine start queue: %s", index, queueName)

			processor := processors[index]
			client := mqClients[index]
			client.ConsumeMessage(c.ctx, consumer, func(msg interface{}) bool {
				c.processPool.Submit(func() {
					// 这里处理消息
					// log.Infof("process msg %v", msg)
					processor.Handle(msg)
				})

				return true
			})

			// client.Close()
		}(i)
	}

	<-c.ctx.Done()
	// 关闭mq连接
	for _, v := range mqClients {
		v.Close()
	}

	// 等待任务完成释放工作池
	c.processPool.ReleaseTimeout(30 * time.Second)
	log.Infof("%s consumer stopped", queueName)
}

func (c *Consumer) newMQClient() []mq.MQ {
	clients := make([]mq.MQ, c.consumerConfig.ConsumerNum)
	server := c.consumerConfig.MQServer
	if server == defaultMQ {
		server = c.config.MQDefault
	}

	for i := 0; i < c.consumerConfig.ConsumerNum; i++ {
		index := i
		clients[index] = client.CreateMQ(server)
	}

	log.Infof("MQServer:%s MQClient num:%d", server, len(clients))

	return clients
}

func (c *Consumer) newProcessor() []process.Processor {
	processors := make([]process.Processor, c.consumerConfig.ConsumerNum)

	pOpts := &process.Options{HTTPOptions: []process.HTTPOptions{process.WithHTTPHost(c.processCfg.Host), process.WithHTTPTimeout(60), process.WithEndpoint(c.consumerConfig.URLPath)}}
	processorType := c.consumerConfig.Processor
	if c.consumerConfig.Processor == process.DefaultProcessorType {
		processorType = c.config.ProcessorDefault
	}
	for i := 0; i < c.consumerConfig.ConsumerNum; i++ {
		processor, err := process.GetProcessor(processorType)
		if err != nil {
			log.Errorf("Failed to get processor: %v", err)
		}
		processor.Init(pOpts)
		processors[i] = processor
	}

	return processors
}

func (c *Consumer) queueName(name string) string {
	if c.config.AppMode != "" {
		return fmt.Sprintf("%s-%s%s", name, c.config.AppMode, c.config.QueueRegion)
	}

	return name
}
