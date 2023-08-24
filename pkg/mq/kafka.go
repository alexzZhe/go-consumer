package mq

import (
	"context"
	"crypto/tls"
	"strings"
	"sync"
	"time"

	"example.com/demo/pkg/log"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
	"github.com/segmentio/kafka-go/snappy"
)

// KafkaConf defines kafka specific options.
type KafkaConf struct {
	Broker                string        `json:"broker" mapstructure:"broker"`
	SSLCertFile           string        `json:"ssl_cert_file" mapstructure:"ssl_cert_file"`
	SSLKeyFile            string        `json:"ssl_key_file" mapstructure:"ssl_key_file"`
	SASLMechanism         string        `json:"sasl_mechanism" mapstructure:"sasl_mechanism"`
	Username              string        `json:"sasl_username" mapstructure:"sasl_username"`
	Password              string        `json:"sasl_password" mapstructure:"sasl_password"`
	Algorithm             string        `json:"sasl_algorithm" mapstructure:"sasl_algorithm"`
	Timeout               time.Duration `json:"timeout" mapstructure:"timeout"`
	Compressed            bool          `json:"compressed" mapstructure:"compressed"`
	UseSSL                bool          `json:"use_ssl" mapstructure:"use_ssl"`
	SSLInsecureSkipVerify bool          `json:"ssl_insecure_skip_verify" mapstructure:"ssl_insecure_skip_verify"`
}

var (
	writer *kafka.Writer
	once   sync.Once
	// reader       sync.Map
	consumeKafka sync.WaitGroup
)

type KafkaMQ struct {
	kafkaDialer *kafka.Dialer
	kafkaConf   *KafkaConf
	writer      *kafka.Writer
	// reader      *kafka.Reader
}

func NewKafkaConf() *KafkaConf {
	return &KafkaConf{
		Broker:  "",
		Timeout: 10 * time.Second,
	}
}

func NewKafkaMQ(cfg *KafkaConf) MQ {

	kafka := &KafkaMQ{kafkaConf: cfg}
	err := kafka.dialer()
	if err != nil {
		log.Errorf("kafka dialer error: %v", err)
	}

	return kafka
}

func (mq *KafkaMQ) dialer() error {
	var tlsConfig *tls.Config
	var err error
	// nolint: nestif
	if mq.kafkaConf.UseSSL {
		if mq.kafkaConf.SSLCertFile != "" && mq.kafkaConf.SSLKeyFile != "" {
			var cert tls.Certificate
			log.Debug("Loading certificates for mTLS.")
			cert, err = tls.LoadX509KeyPair(mq.kafkaConf.SSLCertFile, mq.kafkaConf.SSLKeyFile)
			if err != nil {
				log.Debugf("Error loading mTLS certificates: %s", err.Error())

				return errors.Wrap(err, "failed loading mTLS certificates")
			}
			tlsConfig = &tls.Config{
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: mq.kafkaConf.SSLInsecureSkipVerify,
			}
		} else if mq.kafkaConf.SSLCertFile != "" || mq.kafkaConf.SSLKeyFile != "" {
			log.Error("Only one of ssl_cert_file and ssl_cert_key configuration option is setted, you should set both to enable mTLS.")
		} else {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: mq.kafkaConf.SSLInsecureSkipVerify,
			}
		}
	} else if mq.kafkaConf.SASLMechanism != "" {
		log.Warn("SASL-Mechanism is setted but use_ssl is false.", log.String("SASL-Mechanism", mq.kafkaConf.SASLMechanism))
	}

	var mechanism sasl.Mechanism

	switch mq.kafkaConf.SASLMechanism {
	case "":
		break
	case "PLAIN", "plain":
		mechanism = plain.Mechanism{Username: mq.kafkaConf.Username, Password: mq.kafkaConf.Password}
	case "SCRAM", "scram":
		algorithm := scram.SHA256
		if mq.kafkaConf.Algorithm == "sha-512" || mq.kafkaConf.Algorithm == "SHA-512" {
			algorithm = scram.SHA512
		}
		var mechErr error
		mechanism, mechErr = scram.Mechanism(algorithm, mq.kafkaConf.Username, mq.kafkaConf.Password)
		if mechErr != nil {
			log.Fatalf("Failed initialize kafka mechanism: %s", mechErr.Error())
		}
	default:
		log.Warn(
			"doesn't support this SASL mechanism.",
			log.String("SASL-Mechanism", mq.kafkaConf.SASLMechanism),
		)
	}

	// Kafka connection config
	mq.kafkaDialer = &kafka.Dialer{
		Timeout: mq.kafkaConf.Timeout,
		// ClientID:      mq.kafkaConf.ClientID,
		TLS:           tlsConfig,
		SASLMechanism: mechanism,
	}

	return nil
}

func (mq *KafkaMQ) producer() *kafka.Writer {
	once.Do(func() {
		writerConfig := kafka.WriterConfig{
			Brokers:      strings.Split(mq.kafkaConf.Broker, ","),
			Balancer:     kafka.CRC32Balancer{}, // 使用 kafka.CRC32Balancer 平衡器来获得与 librdkafka 默认的一致性随机分区策略相同的行为。
			Dialer:       mq.kafkaDialer,
			WriteTimeout: mq.kafkaConf.Timeout,
			ReadTimeout:  mq.kafkaConf.Timeout,
			RequiredAcks: 1,
		}
		if mq.kafkaConf.Compressed {
			writerConfig.CompressionCodec = snappy.NewCompressionCodec()
		}

		// w := &kafka.Writer{
		// 	Addr:     kafka.TCP(mq.kafkaConf.Broker...),
		// 	Balancer: &kafka.LeastBytes{},
		// }
		w := kafka.NewWriter(writerConfig)
		writer = w
	})

	return writer
}

func (mq *KafkaMQ) SendMessage(ctx context.Context, queueName string, messages ...*Message) error {
	var err error
	var kMessage []kafka.Message
	for _, msg := range messages {
		kMessage = append(kMessage, kafka.Message{
			Topic: queueName,
			Key:   []byte(msg.Key),
			Value: []byte(msg.Content),
		})

	}
	// log.Infof("kafka send message: %+v", kMessage)
	err = mq.producer().WriteMessages(ctx, kMessage...)
	if err != nil {
		log.Errorf("failed to write messages: %v", err)
	}

	return err
}

func (mq *KafkaMQ) ConsumeMessage(ctx context.Context, c Consumer, callback func(interface{}) bool) error {
	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Dialer:         mq.kafkaDialer,
		Brokers:        strings.Split(mq.kafkaConf.Broker, ","),
		GroupID:        c.GroupID,
		Topic:          c.QueueName,
		MaxBytes:       10e6,        // 10MB
		CommitInterval: time.Second, // flushes commits to Kafka every second
	})
	// kafkaReader := mq.reader
	id := uuid.Must(uuid.NewV4(), nil).String()
	// reader.Store(r, id)

	consumeKafka.Add(1)
	defer consumeKafka.Done()

	var err error
	if c.AutoAck {
		log.Infof("AutoAck Consume start %s", id)
		for {
			m, err := kafkaReader.ReadMessage(ctx)
			if err != nil {
				log.Infof("ReadMessage : %v", err)
				break
			}
			log.Infof("AutoAck message at topic/partition/offset %v/%v/%v: %s = %s id:%s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value), id)
			callback(m.Value)
		}
	} else {
		log.Infof("Consume start %s", id)
		for {
			m, err := kafkaReader.FetchMessage(ctx)
			if err != nil {
				log.Infof("FetchMessage : %v", err)
				break
			}
			log.Infof("message at topic/partition/offset %v/%v/%v: %s = %s id:%s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value), id)
			callback(m.Value)
			if err := kafkaReader.CommitMessages(ctx, m); err != nil {
				log.Errorf("failed to commit messages: %v", err)
			}
		}
	}

	err = kafkaReader.Close()
	if err != nil {
		log.Errorf("failed to close reader: %v", err)
	}
	// reader.Delete(r)
	log.Infof("ConsumeMessage close %s", id)

	return err
}

func (mq *KafkaMQ) Close() error {
	// if mq.producer() != nil {
	// 	if err := mq.producer().Close(); err != nil {
	// 		log.Errorf("kafka writer Close err: %v", err)
	// 	} else {
	// 		log.Infof("kafka writer Close")
	// 	}
	// }

	// reader.Range(func(key, value any) bool {

	// 	err := key.(*kafka.Reader).Close()
	// 	if err != nil {
	// 		log.Errorf("kafka Reader Close err: %v", err)
	// 	}
	// 	log.Infof("kafka reader Close %s", value)
	// 	return true
	// })
	// consumeKafka.Wait()

	if writer != nil {
		err := writer.Close()
		if err != nil {
			log.Errorf("kafka writer Close err: %v", err)
		} else {
			log.Infof("kafka writer Close")
		}
	}

	// if mq.reader != nil {
	// 	err := mq.reader.Close()
	// 	if err != nil {
	// 		log.Errorf("kafka reader Close err: %v", err)
	// 	} else {
	// 		log.Infof("kafka reader Close")
	// 	}
	// }

	consumeKafka.Wait()

	return nil
}
