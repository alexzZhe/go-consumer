package consume

import "example.com/demo/internal/pkg/consumer/process"

func init() {
	process.SetProcessor("php", &ExecPHPProcessor{})
}
