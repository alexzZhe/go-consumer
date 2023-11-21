package php_caller

import (
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestPHPCallerCall(t *testing.T) {
    var (
        num int = 10
        wg  sync.WaitGroup
    )

    // PHP文件路径
    dir, _ := os.Getwd()
    filename := filepath.Join(dir, "demo.php")

    // 开多个协程执行调用
    caller := new(PHPCaller)

    wg.Add(num)

    for i := 0; i < num; i++ {
        // 姓名
        name := fmt.Sprintf("fdipzone_%d", i)

        go func(c *PHPCaller, name string) {
            // 整理请求参数
            params := make([]string, 0, 2)
            params = append(params, "--name", name)

            // 执行请求
            response, err := c.Call(filename, params...)
            assert.Equal(t, nil, err)
            if err != nil {
                return
            }

            assert.Equal(t, fmt.Sprintf("Hello World, %s", name), response.Output)
            wg.Done()
        }(caller, name)
    }

    wg.Wait()
}

func TestPHPCallerCallHelp(t *testing.T) {
    // PHP文件路径
    dir, _ := os.Getwd()
    filename := filepath.Join(dir, "demo.php")

    // 整理请求参数
    params := make([]string, 0, 1)
    params = append(params, "--help")

    // 执行请求
    caller := new(PHPCaller)
    response, err := caller.Call(filename, params...)
    assert.Equal(t, nil, err)
    if err != nil {
        return
    }

    expect_output := fmt.Sprintf("cmd params:\n\n--name Name Parameter\n\n")
    assert.Equal(t, expect_output, response.Output)
}

func TestPHPCallerCallFileNotFound(t *testing.T) {
    // PHP文件路径
    dir, _ := os.Getwd()
    filename := filepath.Join(dir, "not_exists.php")

    // 执行请求
    caller := new(PHPCaller)
    _, err := caller.Call(filename)
    assert.Equal(t, ErrFileNotExists, err)
}
