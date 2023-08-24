package process

var availableProcessors map[string]Processor

func init() {
	availableProcessors = make(map[string]Processor)

	availableProcessors["http"] = &HTTPProcessor{}
}
