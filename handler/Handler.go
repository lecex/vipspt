package handler

import (
	"github.com/micro/go-micro/v2"

	pb "github.com/lecex/pay/proto/tradeService"
	"github.com/lecex/user/core/env"

	"github.com/lecex/vipspt/config"
)

const topic = "event"

// Handler 注册方法
type Handler struct {
	Service micro.Service
}

var Conf = config.Conf

// Register 注册
func (srv *Handler) Register() {
	server := srv.Service.Server()
	pb.RegisterTradesHandler(server, &Trade{
		NotifyUrl:  env.Getenv("PAY_NOTIFY_URL", "http://127.0.01/"),
		PayService: env.Getenv("PAY_SERVICE", "go.micro.srv.pay"),
	})
}
