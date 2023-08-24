package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"sync"

	"example.com/demo/pkg/log"
	"example.com/demo/pkg/mq"
	"example.com/demo/pkg/storage"
)

var kafkaCfg = &mq.KafkaConf{
	Broker: "localhost:9092",
}
var kafkaClient = mq.NewKafkaMQ(kafkaCfg)

// var topicName = "test-queue"
var topicName = "topic-C"

var rabbitmqCfg = &mq.RabbitMQConfig{
	AmqpURI:         "amqp://guest:guest@localhost:5672/",
	ChannelPoolSize: 10,
	ConfirmDelivery: true,
	ExchangeName:    "demo",
}

var redisConfig = &storage.Config{
	Host: "127.0.0.1",
	Port: 6379,
	// Addrs:                 s.redisOptions.Addrs,
	// MasterName:            s.redisOptions.MasterName,
	// Username:              s.redisOptions.Username,
	// Password:              s.redisOptions.Password,
	Database:  0,
	MaxIdle:   10,
	MaxActive: 20,
	Timeout:   1,
	// EnableCluster:         s.redisOptions.EnableCluster,
	// UseSSL:                s.redisOptions.UseSSL,
	// SSLInsecureSkipVerify: s.redisOptions.SSLInsecureSkipVerify,
}

var mnsConfig = &mq.MNSConfig{
	Endpoint:        "",
	AccessKeyID:     "",
	AccessKeySecret: "",
}

var ctxWithCancel, cancel = context.WithCancel(context.Background())

const (
	rabbitmqType = "rabbitmq"
	kafkaType    = "kafka"
	redisType    = "redis"
	mnsType      = "mns"
)

var clientType = rabbitmqType

func main() {

	// client := mq.NewRabbitMQ(rabbitmqCfg)
	// client := mq.NewKafkaMQ(kafkaCfg)
	// client := mq.NewRedisMQ(redisConfig)
	// client := mq.NewAliMNSMQ(mnsConfig)

	// createTopic()
	// listTopic()
	// return

	// Consume(4)
	// return

	client := getClient()

	// make a writer that produces to topic-A, using the least-bytes distribution
	// w := &kafka.Writer{
	// 	Addr: kafka.TCP("localhost:9092"),
	// 	// Topic:    "topic-B",
	// 	Balancer: &kafka.LeastBytes{},
	// }
	var wg sync.WaitGroup
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			defer wg.Done()
			// err := w.WriteMessages(context.Background(),
			// 	kafka.Message{
			// 		Topic: "topic-B",
			// 		Key:   []byte("Key-A"),
			// 		Value: []byte("Hello World!"),
			// 	},
			// 	// kafka.Message{
			// 	// 	Topic: "topic-B",
			// 	// 	Key:   []byte("Key-B" + strconv.Itoa(i)),
			// 	// 	Value: []byte("One!"),
			// 	// },
			// 	// kafka.Message{
			// 	// 	Topic: "topic-B",
			// 	// 	Key:   []byte("Key-C" + strconv.Itoa(i)),
			// 	// 	Value: []byte("Two!"),
			// 	// },
			// )
			// if err != nil {
			// 	log.Fatalf("failed to write messages:", err)
			// }
			// if i < 5 {
			// 	i = 1
			// } else {
			// 	i = 2
			// }
			client := getClient()
			for j := 0; j < 1; j++ {
				SendMessage(client, j)
			}

		}(i)
	}

	wg.Wait()

	client.Close()
	// time.Sleep(20 * time.Second)

}

func getClient() mq.MQ {
	var client mq.MQ
	switch clientType {
	case rabbitmqType:
		client = mq.NewRabbitMQ(rabbitmqCfg)
	case kafkaType:
		client = mq.NewKafkaMQ(kafkaCfg)
	case redisType:
		client = mq.NewRedisMQ(redisConfig)
	case mnsType:
		client = mq.NewAliyunMNS(mnsConfig)
	default:
		panic("client type error")
	}

	return client
}

func SendMessage(pub mq.MQ, i int) {
	msg := mq.Message{
		Key:     fmt.Sprintf("key%d", i),
		Content: "Content",
	}
	pub.SendMessage(context.Background(), topicName, &msg)
}

func kafkaSendMessage(i int) {
	// cfg := mq.KafkaConfig{
	// 	Address: "localhost:9092",
	// }

	msg := mq.Message{
		Key:     fmt.Sprintf("key%d", i),
		Content: "Content",
	}

	err := kafkaClient.SendMessage(context.Background(), topicName, &msg)
	if err != nil {
		log.Fatalf("SendMessage err: %v", err)
	}
}

// func closeKafka() {
// 	// cfg := mq.KafkaConfig{
// 	// 	Address: "localhost:9092",
// 	// }
// 	// kafkaClient := mq.NewKafkaMQ(kafkaCfg)
// 	kafkaClient.Close()
// }

func Consume(num int) {
	ch := make(chan os.Signal, 1)
	var wg sync.WaitGroup
	for i := 0; i < num; i++ {
		wg.Add(1)
		go consume(&wg)
	}

	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	cancel()
	wg.Wait()
	// client.Close()
	close(ch)

	// time.Sleep(30 * time.Second)

}

func callback(msg interface{}) bool {
	log.Infof("callback: %s", msg.(string))
	return true
}

func consume(wg *sync.WaitGroup) {
	defer wg.Done()
	// cfg := mq.KafkaConfig{
	// 	Address: "localhost:9092",
	// 	AutoAck: false,
	// }
	// kafkaClient := mq.NewKafkaMQ(cfg)
	client := getClient()

	consumer := mq.Consumer{
		QueueName: topicName,
		GroupID:   "consumer-group-id",
	}
	err := client.ConsumeMessage(ctxWithCancel, consumer, callback)
	if err != nil {

	}
	client.Close()
}

// func createTopic() {
// 	topic := "topic-C"

// 	conn, err := kafka.Dial("tcp", "localhost:9092")
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	defer conn.Close()

// 	controller, err := conn.Controller()
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	var controllerConn *kafka.Conn
// 	controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	defer controllerConn.Close()

// 	topicConfigs := []kafka.TopicConfig{
// 		{
// 			Topic:             topic,
// 			NumPartitions:     4,
// 			ReplicationFactor: 1,
// 		},
// 	}

// 	err = controllerConn.CreateTopics(topicConfigs...)
// 	if err != nil {
// 		panic(err.Error())
// 	}
// }

// func listTopic() {
// 	conn, err := kafka.Dial("tcp", "localhost:9092")
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	defer conn.Close()

// 	partitions, err := conn.ReadPartitions()
// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	m := map[string]struct{}{}

// 	for _, p := range partitions {
// 		m[p.Topic] = struct{}{}
// 	}
// 	for k := range m {
// 		fmt.Println(k)
// 	}
// }
