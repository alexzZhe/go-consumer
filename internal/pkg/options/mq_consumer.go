package options

// MQConsumerConfig defines options for mq consume.
type MQConsumerConfig struct {
	Group               string `json:"group"   mapstructure:"group"`
	QueueName           string `json:"queue-name"   mapstructure:"queue-name"`
	Path                string `json:"path"   mapstructure:"path"`
	ConsumerNum         int    `json:"consumer-num"   mapstructure:"consumer-num"`
	MessageProcessorNum int    `json:"message-processor-num"   mapstructure:"message-processor-num"`
	GroupID             string `json:"group-id"   mapstructure:"group-id"`
	MQServer            string `json:"mq-server"   mapstructure:"mq-server"`
	Processor           string `json:"processor" mapstructure:"processor"`
}
