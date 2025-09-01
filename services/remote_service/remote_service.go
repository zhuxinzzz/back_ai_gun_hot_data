package remote_service

import (
	"back_ai_gun_data/pkg/lr"
	"time"

	"github.com/go-resty/resty/v2"
)

func GetHost() string {
	//return "https://api.aigun.dev"
	return "http://192.168.2.18:12345"
}

// 扁平化HTTP客户端 - 直接使用，不包装
var cli *resty.Client

func Init() {
	cli = resty.New().
		SetTimeout(30 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second)
	//cli.SetDebug(true)
	// 设置通用请求头
	cli.SetHeader("User-Agent", "back_ai_gun_data/1.0")
	cli.SetHeader("Content-Type", "application/json")

	// 设置响应处理
	cli.OnAfterResponse(func(client *resty.Client, resp *resty.Response) error {
		if resp.StatusCode() >= 400 {
			lr.E().Errorf("HTTP error: %d - %s", resp.StatusCode(), resp.String())
		}
		return nil
	})

}

func Cli() *resty.Client {
	return cli
}
