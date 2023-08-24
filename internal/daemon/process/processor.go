package process

import "fmt"

const (
	HTTPProcessorType    = "http"
	DefaultProcessorType = "default"
)

type HTTPOptions func(p *HTTPProcessor)

type Options struct {
	HTTPOptions []HTTPOptions
}

type Processor interface {
	New() Processor
	Init(opts *Options)
	Handle(v any) error
}

func GetProcessor(t string) (Processor, error) {
	if p, ok := availableProcessors[t]; ok {
		return p.New(), nil
	}
	return nil, fmt.Errorf("processor %s not found", t)
}
