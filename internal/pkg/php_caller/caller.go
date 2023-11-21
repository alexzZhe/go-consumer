// 利用os/exec包执行调用 PHP 文件
// exec.Command("php", filename) 是使用Go的os/exec包来执行命令行中的php命令，并指定要调用的PHP文件。
// 它实际上是在操作系统级别启动一个新的进程来运行PHP解释器，并将指定的PHP文件作为参数传递给解释器。
// 这种方式适用于需要在独立的进程中运行PHP代码的情况，例如执行外部命令、调用外部API等。
package php_caller

import (
    "os"
    "os/exec"
)

// 定义 PHPCaller 必须实现 ICaller 接口
var _ ICaller = &PHPCaller{}

// PHPCaller 定义PHP文件执行器结构
type PHPCaller struct {
}

// Call 执行调用 PHP 文件
// filename 要执行的 PHP 文件
// params 执行 PHP 文件所需参数
// response PHP 文件执行结果
func (pc *PHPCaller) Call(filename string, params ...string) (CallerResponse, error) {
    // 判断文件是否存在
    _, err := os.Stat(filename)
    if err != nil {
        return CallerResponse{}, ErrFileNotExists
    }

    // 整理 args
    var args []string = make([]string, 0, 4)
    args = append(args, filename)

    for _, param := range params {
        args = append(args, param)
    }

    // 执行命令
    cmd := exec.Command("php", args...)
    output, err := cmd.Output()

    if err != nil {
        return CallerResponse{}, err
    }

    // 输出
    return CallerResponse{
        Output: string(output),
    }, nil
}
