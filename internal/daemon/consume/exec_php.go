package consume

import (
	"example.com/demo/internal/pkg/consumer/process"
	"example.com/demo/internal/pkg/php_caller"
	"example.com/demo/pkg/log"
)

// exec php
type ExecPHPProcessor struct {
	path   string
	caller *php_caller.PHPCaller
}

func (p *ExecPHPProcessor) New() process.Processor {
	newProcess := &ExecPHPProcessor{}

	return newProcess
}

func (p *ExecPHPProcessor) Init(opts *process.Options) {
	p.path = opts.Path
	p.caller = new(php_caller.PHPCaller)
}

func (p *ExecPHPProcessor) Handle(v any) error {
	log.Infof("ExecPHPProcessor %v", string(v.([]byte)))

	params := []string{"--name", string(v.([]byte))}
	resp, err := p.caller.Call(p.path, params...)
	if err != nil {
		log.Errorf("Call php %v %v", p.path, err.Error())
	}
	log.Infof("Call resp: %v", resp.Output)

	return err
}
