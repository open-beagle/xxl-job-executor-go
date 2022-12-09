package main

import (
	"fmt"
	"log"

	xxl "github.com/MasterYang7/xxl-job-executor-go"
	"github.com/MasterYang7/xxl-job-executor-go/example/task"
)

func main() {
	exec := xxl.NewExecutor(
		xxl.ServerAddr("http://localhost:1184/xxl-job-admin"),
		xxl.AccessToken("default_token"),       //请求令牌(默认为空)
		xxl.ExecutorIp("10.3.208.81"),          //可自动获取
		xxl.ExecutorPort("9999"),               //默认9999（非必填）
		xxl.RegistryKey("gsy-golang-jobs-001"), //执行器名称
		xxl.SetRegistryAlias("gsy测试执行器"),       // 设置别名
		xxl.SetLogger(&logger{}),               //自定义日志
		xxl.SetAdminPwd("123456"),              // 超管密码
	)
	exec.Init()
	//设置日志查看handler
	exec.LogHandler(func(req *xxl.LogReq) *xxl.LogRes {
		return &xxl.LogRes{Code: xxl.SuccessCode, Msg: "", Content: xxl.LogResContent{
			FromLineNum: req.FromLineNum,
			ToLineNum:   2,
			LogContent:  "这个是自定义日志handler",
			IsEnd:       true,
		}}
	})
	//注册任务handler
	exec.RegTask("task.test-001", "描述1", "0/1 * * * * ?", task.Test)
	log.Fatal(exec.Run())
}

// xxl.Logger接口实现
type logger struct{}

func (l *logger) Info(format string, a ...interface{}) {
	fmt.Println(fmt.Sprintf("【XXL-JOB】日志 - "+format, a...))
}

func (l *logger) Error(format string, a ...interface{}) {
	log.Println(fmt.Sprintf("【XXL-JOB】日志 - "+format, a...))
}
