package main

import (
	"back_ai_gun_data/pkg/dao"
	"back_ai_gun_data/services/remote_service"
	"back_ai_gun_data/utils"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"back_ai_gun_data/consumer"
	"back_ai_gun_data/pkg/cache"
	"back_ai_gun_data/pkg/lr"

	"github.com/sirupsen/logrus"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("panic: %v", r)
			lr.E().WithFields(logrus.Fields{
				"backtrace": utils.GetStack(),
			}).Error(msg)
		}
	}()

	lr.Init()
	dao.Init()
	cache.Init()
	remote_service.Init()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer.StartAllConsumers(ctx)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	cancel()
}
