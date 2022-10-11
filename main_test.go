package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	pb "github.com/lecex/pay/proto/tradeService"
	"github.com/lecex/vipspt/handler"
)

var Config = map[string]string{
	"Appid":         os.Getenv("PAY_VIPSPT_APPID"),
	"SecretKey":     os.Getenv("PAY_VIPSPT_SECRET_KEY"),
	"SubMerId":      os.Getenv("PAY_VIPSPT_MERCHANT_IDD"),
	"EnterpriseReg": os.Getenv("PAY_VIPSPT_ENTERPRISE_REG"),
	"Sandbox":       fmt.Sprintf("%t", false),
	// "Order":               `{"created_at":"2022-09-08T15:04:05+08:00"}`,
	// "RefundOrder":         `{"created_at":"2022-09-08T15:04:05+08:00"}`,
}

func TestAopF2F(t *testing.T) {
	req := &pb.Request{
		Config: Config,
		BizContent: &pb.BizContent{
			Method:     "wechat",
			OutTradeNo: "513457061273811890",
			TotalFee:   "1",
			Title:      "测试商品",
			AuthCode:   "131144114460498167",
		},
	}
	res := &pb.Response{}
	h := &handler.Trade{
		NotifyUrl: "http://127.0.0.1:8080/notify",
	}
	err := h.AopF2F(context.TODO(), req, res)
	fmt.Println("TestAopF2F", res, err)
	t.Log(req, res, err)
}

//  51345706127381181888
func TestPayQuery(t *testing.T) {
	// req := &pb.Request{
	// 	Config: Config,
	// 	BizContent: &pb.BizContent{
	// 		OutTradeNo: "51345706127381181888",
	// 	},
	// }
	// res := &pb.Response{}
	// h := &handler.Trade{}
	// err := h.Query(context.TODO(), req, res)
	// fmt.Println("TestQuery", res, err)
	// t.Log(req, res, err)
}

func TestPayRefund(t *testing.T) {
	// req := &pb.Request{
	// 	Config: Config,
	// 	BizContent: &pb.BizContent{
	// 		OutRefundNo: "51345706127381181888-1",
	// 		RefundFee:   "1",
	// 		OutTradeNo:  "51345706127381181888",
	// 	},
	// }
	// res := &pb.Response{}
	// h := &handler.Trade{}
	// err := h.Refund(context.TODO(), req, res)
	// fmt.Println("TestRefund", res, err)
	// t.Log(req, res, err)
}

func TestPayRefundQuery(t *testing.T) {
	// 创建连接
	// req := &pb.Request{
	// 	Config: Config,
	// 	BizContent: &pb.BizContent{
	// 		OutRefundNo: "51345706127381181888-1",
	// 		OutTradeNo:  "51345706127381181888",
	// 	},
	// }
	// res := &pb.Response{}
	// h := &handler.Trade{}
	// err := h.RefundQuery(context.TODO(), req, res)
	// fmt.Println("TestRefundQuery", res, err)
	// t.Log(req, res, err)
}

func TestJsApi(t *testing.T) {
	// 创建连接
	// req := &pb.Request{
	// 	Config: Config,
	// 	BizContent: &pb.BizContent{
	// 		Method:     "alipay",
	// 		OutTradeNo: "151345706127381181883",
	// 		TotalFee:   "1",
	// 		Title:      "测试商品",
	// 		OpenId:     "2088002104076813",
	// 		// OpenId:     "okCtS6IyyODgL6EyAI3HQLUEN-cs",
	// 		// AppId:      "wx26e296a18096b757",
	// 	},
	// }
	// res := &pb.Response{}
	// h := &handler.Trade{}
	// err := h.JsApi(context.TODO(), req, res)
	// fmt.Println("TestJsApi", res, err)
	// t.Log(req, res, err)
}

func TestPayOpenId(t *testing.T) {
	// req := &pb.Request{
	// 	Config: Config,
	// 	BizContent: &pb.BizContent{
	// 		Method:   "wechat",
	// 		AuthCode: "130768632451713032",
	// 		AppId:    "wx26e296a18096b757",
	// 	},
	// }
	// res := &pb.Response{}
	// h := &handler.Trade{}
	// err := h.OpenId(context.TODO(), req, res)
	// fmt.Println("TestOpenId", res, err)
	// t.Log(req, res, err)
}
