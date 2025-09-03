package remote_service

import (
	"back_ai_gun_data/pkg/lr"
	"time"

	"github.com/go-resty/resty/v2"
)

func GetHost() string {
	return "http://api.goldriver.xyz.satoshi8.world"
	//return "http://192.168.2.18:12345"
}

var cli *resty.Client

func Init() {
	cli = resty.New().
		//SetTimeout(30 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second)
	//cli.SetDebug(true)

	cli.SetHeader("User-Agent", "back_ai_gun_data/1.0")
	cli.SetHeader("Content-Type", "application/json")

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
