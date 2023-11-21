package process

import (
	"fmt"
)

const (
	HTTPProcessorType    = "http"
	DefaultProcessorType = "default"
)

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

func SetProcessor(t string, p Processor) error {
	if _, ok := availableProcessors[t]; ok {
		return fmt.Errorf("processor %s already exists", t)
	}
	availableProcessors[t] = p

	return nil
}
