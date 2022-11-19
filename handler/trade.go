package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/clbanning/mxj"
	tradePB "github.com/lecex/pay/proto/trade"
	pb "github.com/lecex/pay/proto/tradeService"
	client "github.com/lecex/user/core/client"
	"github.com/shopspring/decimal"

	"github.com/lecex/vipspt/service"
	"github.com/lecex/vipspt/service/requests"
)

// Trade 支付结构
type Trade struct {
	NotifyUrl  string
	PayService string
}

// 初始化链接
func (srv *Trade) NewClient(config map[string]string) (client *service.Client, err error) {
	if v, ok := config["Appid"]; !ok && v == "" {
		return nil, fmt.Errorf("vipspt Appid is empty")
	}
	if v, ok := config["SecretKey"]; !ok && v == "" {
		return nil, fmt.Errorf("vipspt SecretKey is empty")
	}
	if v, ok := config["SubMerId"]; !ok && v == "" {
		return nil, fmt.Errorf("vipspt SubMerId is empty")
	}
	if v, ok := config["EnterpriseReg"]; !ok && v == "" {
		return nil, fmt.Errorf("vipspt EnterpriseReg is empty")
	}
	if v, ok := config["Sandbox"]; !ok && v == "" {
		return nil, fmt.Errorf("vipspt Sandbox is empty")
	}

	sandbox, _ := strconv.ParseBool(config["Sandbox"])
	client = service.NewClient()
	client.Config.Appid = config["Appid"]
	client.Config.SecretKey = config["SecretKey"]
	client.Config.MerchantId = config["SubMerId"]
	client.Config.EnterpriseReg = config["EnterpriseReg"]
	if v, ok := config["NotifyUrl"]; ok {
		client.Config.NotifyUrl = v
	}
	client.Config.Sandbox = sandbox
	return client, nil
}

// request 请求处理
func (srv *Trade) request(request *requests.CommonRequest, req *pb.Request, res *pb.Response) (err error) {
	client, err := srv.NewClient(req.Config)
	if err != nil {
		return err
	}
	response, err := client.ProcessCommonRequest(request)
	if err != nil {
		return err
	}
	data, err := response.GetVerifySignDataMap()
	if err != nil {
		return err
	}
	r, err := data.Json()
	if err != nil {
		return err
	}
	res.Content = string(r)
	return err
}

func (srv *Trade) Notify(ctx context.Context, req *pb.NotifyRequest, res *pb.NotifyResponse) (err error) {
	get, post, _ := srv.handlerRequest(req)
	if v, ok := get["id"]; !ok || v == nil {
		return fmt.Errorf("未找到id参数")
	}
	// to json string
	postJson, err := post.Json()
	if err != nil {
		return err
	}
	r := &tradePB.NotifyRequest{
		Id:         get["id"].(string),
		NotifyData: string(postJson),
	}
	rs := &tradePB.NotifyResponse{}
	err = client.Call(ctx, srv.PayService, "Trades.Notify", r, rs)
	if err != nil {
		return err
	}
	if rs.ReturnCode == "SUCCESS" {
		res.StatusCode = http.StatusOK
		res.Body = "success"
	} else {
		res.StatusCode = http.StatusOK
		res.Body = "FAIL"
	}
	return nil
}

func (srv *Trade) HanderNotify(ctx context.Context, req *pb.Request, res *pb.Response) (err error) {
	return fmt.Errorf("暂不支持,HanderNotify:vipspt")
}

func (srv *Trade) AopF2F(ctx context.Context, req *pb.Request, res *pb.Response) (err error) {
	payWay := "WXZF"
	// 配置参数
	switch req.BizContent.Method {
	case "wechat":
		payWay = "WXZF"
	case "alipay":
		payWay = "ZFBZF"
	default:
		return fmt.Errorf("暂不支持," + req.BizContent.Method + ":vipspt")
	}
	totalFee, err := strconv.ParseFloat(req.BizContent.TotalFee, 64)
	if err != nil {
		return err
	}
	request := requests.NewCommonRequest()
	request.ApiName = "pay.pay"
	request.BizContent = map[string]interface{}{
		"merchant_id":   req.Config["SubMerId"],
		"enterpriseReg": req.Config["EnterpriseReg"],
		"pay_way":       payWay,
		"out_order_id":  req.BizContent.OutTradeNo,                                              // 商户订单号(商户交易系统中唯一)
		"sMchtIp":       "127.0.0.1",                                                            // 业务代码
		"sAuthCode":     req.BizContent.AuthCode,                                                // 商品名称
		"amount":        decimal.NewFromFloat(totalFee).Div(decimal.NewFromFloat(float64(100))), // 单位为分
		// 交易时间 date_time:2021-06-22 13:48:55
		"date_time": time.Now().Format("2006-01-02 15:04:05"),
	}
	return srv.request(request, req, res)
}

func (srv *Trade) Query(ctx context.Context, req *pb.Request, res *pb.Response) (err error) {
	// 配置参数
	order, err := mxj.NewMapJson([]byte(req.Config["Order"]))
	if err != nil {
		return err
	}
	request := requests.NewCommonRequest()
	request.ApiName = "pay.query"
	request.BizContent = map[string]interface{}{
		"merchant_id":   req.Config["SubMerId"],
		"enterpriseReg": req.Config["EnterpriseReg"],
	}
	if v, ok := order["bank_trade_no"]; ok && v != nil && v != "" {
		request.BizContent["third_order_id"] = v
	} else {
		request.BizContent["out_order_id"] = req.BizContent.OutTradeNo
	}
	return srv.request(request, req, res)
}

func (srv *Trade) Refund(ctx context.Context, req *pb.Request, res *pb.Response) (err error) {
	// 配置参数
	originalOrder, err := mxj.NewMapJson([]byte(req.Config["OriginalOrder"]))
	if err != nil {
		return err
	}
	refundFee, err := strconv.ParseFloat(req.BizContent.RefundFee, 64)
	if err != nil {
		return err
	}
	refundMsg := "退款"
	if req.BizContent.Title != "" {
		refundMsg = req.BizContent.Title
	}
	request := requests.NewCommonRequest()
	request.ApiName = "pay.refund"
	request.BizContent = map[string]interface{}{
		"merchant_id":    req.Config["SubMerId"],
		"enterpriseReg":  req.Config["EnterpriseReg"],
		"out_order_id":   req.BizContent.OutRefundNo,
		"third_order_id": originalOrder["bank_trade_no"],
		"refundMsg":      refundMsg,
		"refund_amount":  decimal.NewFromFloat(refundFee).Div(decimal.NewFromFloat(float64(100))),
	}
	return srv.request(request, req, res)
}

func (srv *Trade) RefundQuery(ctx context.Context, req *pb.Request, res *pb.Response) (err error) {
	// 配置参数
	order, err := mxj.NewMapJson([]byte(req.Config["Order"]))
	if err != nil {
		return err
	}
	request := requests.NewCommonRequest()
	request.ApiName = "pay.refundQuery"
	request.BizContent = map[string]interface{}{
		"merchant_id":   req.Config["SubMerId"],
		"enterpriseReg": req.Config["EnterpriseReg"],
	}
	if v, ok := order["bank_trade_no"]; ok && v != nil && v != "" {
		request.BizContent["third_order_id"] = v
	} else {
		request.BizContent["out_order_id"] = req.BizContent.OutRefundNo
	}
	return srv.request(request, req, res)
}

func (srv *Trade) JsApi(ctx context.Context, req *pb.Request, res *pb.Response) (err error) {
	return fmt.Errorf("暂不支持,JsApi:vipspt")
}

// QRCode 构建自己的聚合支付
func (srv *Trade) QRCode(ctx context.Context, req *pb.Request, res *pb.Response) (err error) {
	data := mxj.New()
	data["return_code"] = "SUCCESS"
	data["return_msg"] = "SUCCESS"
	data["qr_code"] = "self"
	content, err := data.Json()
	if err != nil {
		return err
	}
	res.Content = string(content)
	return err
}

func (srv *Trade) OpenId(ctx context.Context, req *pb.Request, res *pb.Response) (err error) {
	return fmt.Errorf("暂不支持,OpenId:vipspt")
}

func (srv *Trade) WxFacePayInfo(ctx context.Context, req *pb.Request, res *pb.Response) (err error) {
	return fmt.Errorf("暂不支持,微信刷脸:vipspt")
}

// handlerRequest 处理请求
func (srv *Trade) handlerRequest(req *pb.NotifyRequest) (get mxj.Map, post mxj.Map, header mxj.Map) {
	get = mxj.New()
	for k, v := range req.Get {
		get[k] = v.Values
	}
	post = mxj.New()
	for k, v := range req.Post {
		post[k] = v.Values
	}
	header = mxj.New()
	for k, v := range req.Header {
		header[k] = v.Values
	}
	return get, post, header
}
