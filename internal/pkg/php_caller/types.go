package php_caller

import "errors"

// 预定义错误类型
var (
    ErrFileNotExists = errors.New("php file not exists")
)

// ICaller Caller接口定义
type ICaller interface {
    Call(filename string, params ...string) (CallerResponse, error)
}

// CallerResponse 定义执行器输出结构
type CallerResponse struct {
    Output string
}
