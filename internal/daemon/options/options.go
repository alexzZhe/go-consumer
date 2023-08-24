// Package options contains flags and options for initializing an server
package options

import (
	cliflag "github.com/marmotedu/component-base/pkg/cli/flag"
	"github.com/marmotedu/component-base/pkg/json"

	genericoptions "example.com/demo/internal/pkg/options"
	"example.com/demo/pkg/log"
	"example.com/demo/pkg/mq"
)

type MQConsumerConfig struct {
	Group               string `json:"group"   mapstructure:"group"`
	QueueName           string `json:"queue-name"   mapstructure:"queue-name"`
	URLPath             string `json:"url-path"   mapstructure:"url-path"`
	ConsumerNum         int    `json:"consumer-num"   mapstructure:"consumer-num"`
	MessageProcessorNum int    `json:"message-processor-num"   mapstructure:"message-processor-num"`
	GroupID             string `json:"group-id"   mapstructure:"group-id"`
	MQServer            string `json:"mq-server"   mapstructure:"mq-server"`
	Processor           string `json:"processor" mapstructure:"processor"`
}

// Options runs a server.
type Options struct {
	AppMode            string                       `json:"app-mode"                mapstructure:"app_mode"`
	QueueRegion        string                       `json:"queue-region"            mapstructure:"queue_region"`
	HealthCheckPath    string                       `json:"health-check-path"       mapstructure:"health-check-path"`
	HealthCheckAddress string                       `json:"health-check-address"    mapstructure:"health-check-address"`
	RedisOptions       *genericoptions.RedisOptions `json:"redis"                   mapstructure:"redis"`
	KafkaConfig        *mq.KafkaConf                `json:"kafka"            mapstructure:"kafka"`
	RabbitMQConfig     *mq.RabbitMQConfig           `json:"rabbitmq"         mapstructure:"rabbitmq"`
	AliyunMNSConfig    *mq.MNSConfig                `json:"aliyunmns"              mapstructure:"aliyunmns"`
	Log                *log.Options                 `json:"log"                     mapstructure:"log"`
	MQConsumer         map[string]MQConsumerConfig  `json:"mq-consumer"      mapstructure:"mq-consumer"`
	MessageProcessHost string                       `json:"message-process-host" mapstructure:"message-process-host"`
	ConsumerGroups     string                       `json:"consumer-groups" mapstructure:"consumer-groups"`
	MQDefault          string                       `json:"mq-default" mapstructure:"mq-default"`
	ProcessorDefault   string                       `json:"processor-default" mapstructure:"processor-default"`
}

// NewOptions creates a new Options object with default parameters.
func NewOptions() *Options {
	s := Options{
		AppMode:            "dev",
		QueueRegion:        "",
		HealthCheckPath:    "health",
		HealthCheckAddress: "0.0.0.0:7070",
		RedisOptions:       genericoptions.NewRedisOptions(),
		KafkaConfig:        mq.NewKafkaConf(),
		RabbitMQConfig:     mq.NewRabbitMQConfig(),
		AliyunMNSConfig:    mq.NewMNSConfig(),
		Log:                log.NewOptions(),
		MQConsumer:         make(map[string]MQConsumerConfig),
		MessageProcessHost: "http://localhost",
		ConsumerGroups:     "",
		MQDefault:          "",
		ProcessorDefault:   "",
	}

	return &s
}

// Flags returns flags for a specific APIServer by section name.
func (o *Options) Flags() (fss cliflag.NamedFlagSets) {
	o.RedisOptions.AddFlags(fss.FlagSet("redis"))
	o.Log.AddFlags(fss.FlagSet("logs"))

	// Note: the weird ""+ in below lines seems to be the only way to get gofmt to
	// arrange these text blocks sensibly. Grrr.
	fs := fss.FlagSet("misc")
	// fs.IntVar(&o.PurgeDelay, "purge-delay", o.PurgeDelay, ""+
	// 	"This setting the purge delay (in seconds) when purge the data from Redis to MongoDB or other data stores.")
	fs.StringVar(&o.HealthCheckPath, "health-check-path", o.HealthCheckPath, ""+
		"Specifies liveness health check request path.")
	fs.StringVar(&o.HealthCheckAddress, "health-check-address", o.HealthCheckAddress, ""+
		"Specifies liveness health check bind address.")
	// fs.BoolVar(&o.OmitDetailedRecording, "omit-detailed-recording", o.OmitDetailedRecording, ""+
	// 	"Setting this to true will avoid writing policy fields for each authorization request in pumps.")

	return fss
}

func (o *Options) String() string {
	data, _ := json.Marshal(o)

	return string(data)
}
