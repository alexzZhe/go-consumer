package daemon

import (
	"context"
	"strings"
	"sync"

	"example.com/demo/pkg/log"

	"example.com/demo/internal/daemon/config"
	"example.com/demo/internal/pkg/client"

	"example.com/demo/pkg/shutdown"
	"example.com/demo/pkg/shutdown/shutdownmanagers/posixsignal"
)

type daemonServer struct {
	gs               *shutdown.GracefulShutdown
	consumerStopFunc context.CancelFunc
	// secInterval    int
	// omitDetails    bool
	// mutex          *redsync.Mutex
	config         *config.Config
	consumerGroups []string
}

// preparedGenericAPIServer is a private wrapper that enforces a call of PrepareRun() before Run can be invoked.
type preparedDaemonServer struct {
	*daemonServer
}

func createDaemonServer(cfg *config.Config) (*daemonServer, error) {
	gs := shutdown.New()
	gs.AddShutdownManager(posixsignal.NewPosixSignalManager())

	// cfg.ConsumerGroups

	server := &daemonServer{
		gs:             gs,
		config:         cfg,
		consumerGroups: strings.Split(cfg.ConsumerGroups, ","),
	}

	return server, nil
}

func (s *daemonServer) PrepareRun() preparedDaemonServer {
	s.initialize()

	return preparedDaemonServer{s}
}

func (s *daemonServer) initialize() {
	// MQ配置初始化
	client.InitMQOptions(client.WithAliyunMNSConfig(s.config.AliyunMNSConfig),
		client.WithKafkaConfig(s.config.KafkaConfig),
		client.WithRabbitMQConfig(s.config.RabbitMQConfig),
		client.WithRedisConfig(s.config.RedisOptions))
}

func (s preparedDaemonServer) Run() error {
	stopCh := make(chan struct{})
	s.gs.AddShutdownCallback(shutdown.ShutdownFunc(func(string) error {

		s.consumerStopFunc()
		stopCh <- struct{}{}

		return nil
	}))

	// start shutdown managers
	if err := s.gs.Start(); err != nil {
		log.Fatalf("start shutdown manager failed: %s", err.Error())
	}

	// 开始 daemon 出队
	s.start()

	// blocking here via channel to prevents the process exit.
	<-stopCh
	return nil
}

func (s *daemonServer) start() {
	ctx, cancel := context.WithCancel(context.Background())
	s.consumerStopFunc = cancel
	var wg sync.WaitGroup

	for key, consumeConfig := range s.config.MQConsumer {
		if !s.filterConsumerGroup(consumeConfig.Group) {
			continue
		}
		wg.Add(1)
		log.Infof("daemon start consumer: %s %+v", key, consumeConfig)

		consumer := NewConsumer(ctx, s.config)
		consumer.Init(consumeConfig)
		go consumer.Start(&wg)
	}

	wg.Wait()
}

func (s *daemonServer) filterConsumerGroup(group string) bool {
	for _, v := range s.consumerGroups {
		if v == group {
			return true
		}
	}

	return false
}
