<?php
// 获取命令行参数
$cmd_params = getopt('', array('name:', 'help::'));

if(isset($cmd_params['help'])){
    echo 'cmd params:'.PHP_EOL.PHP_EOL.'--name Name Parameter'.PHP_EOL.PHP_EOL;
    exit();
}

// 姓名参数
$name = isset($cmd_params['name'])? $cmd_params['name'] : '';

// Demo
Class Demo
{
    public function HelloWorld(string $name):string
    {
        if(empty($name)){
            return sprintf('Hello World');
        }
        return sprintf('Hello World, %s', $name);
    }
}

$o = new Demo;
echo $o->HelloWorld($name);