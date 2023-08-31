package process

import (
	"context"
	"time"

	"example.com/demo/pkg/log"
	gohttpclient "github.com/bozd4g/go-http-client"
)

type HTTPProcessor struct {
	host     string
	endpoint string
	timeout  int
	client   *gohttpclient.Client
}

func (p *HTTPProcessor) New() Processor {
	newProcess := &HTTPProcessor{}

	return newProcess
}

func (p *HTTPProcessor) Init(opts *Options) {
	for _, opt := range opts.HTTPOptions {
		opt(p)
	}

	t := time.Duration(p.timeout) * time.Second
	p.client = gohttpclient.New(p.host, gohttpclient.WithTimeout(t), gohttpclient.WithDefaultHeaders())
}

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

func (p *HTTPProcessor) Handle(v any) error {

	resp, err := p.client.Post(context.Background(), p.endpoint, gohttpclient.WithBody(v.([]byte)))
	if err != nil {
		log.Errorf("Failed to Post: %v", err)
		return err
	}
	if !resp.Ok() {
		log.Errorf("Response: %v %s %s", resp.Status(), p.endpoint, resp.Body())
	} else {
		log.Infof("Response: %v %s %s", resp.Status(), p.endpoint, resp.Body())
	}

	return nil
}
