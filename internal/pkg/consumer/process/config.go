package process

type HTTPOptions func(p *HTTPProcessor)
type ProcessOptions func(p *Options)

type Options struct {
	HTTPOptions []HTTPOptions
	Path        string
}

type ProcessConfig struct {
	Host             string
	MQDefault        string
	ProcessorDefault string
	AppMode          string
	QueueRegion      string
}

type ProcessConfigOption func(*ProcessConfig)

func NewProcessConfig(opts ...ProcessConfigOption) *ProcessConfig {
	processCfg := &ProcessConfig{}

	for _, opt := range opts {
		opt(processCfg)
	}

	return processCfg
}

func WithHost(host string) ProcessConfigOption {
	return func(pc *ProcessConfig) {
		pc.Host = host
	}
}

func WithDefaultMQ(mq string) ProcessConfigOption {
	return func(pc *ProcessConfig) {
		pc.MQDefault = mq
	}
}

func WithDefaultProcessor(processor string) ProcessConfigOption {
	return func(pc *ProcessConfig) {
		pc.ProcessorDefault = processor
	}
}

func WithAppMode(appMode string) ProcessConfigOption {
	return func(pc *ProcessConfig) {
		pc.AppMode = appMode
	}
}

func WithQueueRegion(queueRegion string) ProcessConfigOption {
	return func(pc *ProcessConfig) {
		pc.QueueRegion = queueRegion
	}
}

// func WithHTTPOptions(opt HTTPOptions) ProcessOptions {
// 	return func(p *Options) {
// 		p.HTTPOptions = opt
// 	}
// }

// func WithPath(path string) ProcessOptions {
// 	return func(p *Options) {
// 		p.Path = path
// 	}
// }

func WithHTTPHost(host string) HTTPOptions {
	return func(p *HTTPProcessor) {
		p.host = host
	}
}

func WithHTTPTimeout(timeout int) HTTPOptions {
	return func(p *HTTPProcessor) {
		p.timeout = timeout
	}
}
func WithEndpoint(endpoint string) HTTPOptions {
	return func(p *HTTPProcessor) {
		p.endpoint = endpoint
	}
}
