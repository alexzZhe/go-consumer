package process

import "example.com/demo/pkg/log"

// for test Processor
type DummyProcessor struct {
}

func (p *DummyProcessor) New() Processor {
	newProcess := &DummyProcessor{}

	return newProcess
}

func (p *DummyProcessor) Init(opts *Options) {

}

func (p *DummyProcessor) Handle(v any) error {
	log.Infof("DummyProcessor %v", string(v.([]byte)))
	return nil
}
