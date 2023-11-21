package consumer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"example.com/demo/internal/pkg/client"
	"example.com/demo/internal/pkg/consumer/process"
	"example.com/demo/internal/pkg/options"
	"example.com/demo/pkg/log"
	"example.com/demo/pkg/mq"
	"github.com/panjf2000/ants/v2"
)

type Consumer struct {
	processCfg     *process.ProcessConfig
	processPool    *ants.Pool
	consumerConfig options.MQConsumerConfig
}

const (
	defaultMQ = "default"
)

func New(processCfg *process.ProcessConfig) *Consumer {
	return &Consumer{processCfg: processCfg}
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
func (c *Consumer) Start(ctx context.Context, wg *sync.WaitGroup) {
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
			client.ConsumeMessage(ctx, consumer, func(msg interface{}) bool {
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

	<-ctx.Done()
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
		server = c.processCfg.MQDefault
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

	pOpts := &process.Options{
		HTTPOptions: []process.HTTPOptions{process.WithHTTPHost(c.processCfg.Host), process.WithHTTPTimeout(60), process.WithEndpoint(c.consumerConfig.Path)},
		Path:        c.consumerConfig.Path,
	}
	processorType := c.consumerConfig.Processor
	if c.consumerConfig.Processor == process.DefaultProcessorType {
		processorType = c.processCfg.ProcessorDefault
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
	if c.processCfg.AppMode != "" {
		return fmt.Sprintf("%s-%s%s", name, c.processCfg.AppMode, c.processCfg.QueueRegion)
	}

	return name
}
