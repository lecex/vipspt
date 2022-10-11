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
	"github.com/lecex/vipspt/service/responses"
	"github.com/lecex/vipspt/service/util"
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
	r := &tradePB.Request{
		BizContent: &tradePB.BizContent{
			Id:         get["id"].(string),
			NotifyData: string(postJson),
		},
	}
	rs := &tradePB.Response{}
	err = client.Call(ctx, srv.PayService, "Trades.Notify", r, rs)
	if err != nil {
		return err
	}
	if rs.Content.ReturnCode == "SUCCESS" {
		res.StatusCode = http.StatusOK
		res.Body = "success"
	} else {
		res.StatusCode = http.StatusOK
		res.Body = "FAIL"
	}
	return nil
}

func (srv *Trade) HanderNotify(ctx context.Context, req *pb.Request, res *pb.Response) (err error) {
	content, err := mxj.NewMapJson([]byte(req.BizContent.NotifyData))
	if err != nil {
		return err
	}
	ok, err := util.VerifySign(util.EncodeSignParams(content), content["sign"].(string), "", req.Config["VipsptPublicKeyData"], "RSA")
	if err != nil {
		return err
	}
	if ok {
		r := &responses.CommonResponse{}
		data := r.HanderVipsptNotify(content)
		if err != nil {
			return err
		}
		dataJson, err := data.Json()
		if err != nil {
			return err
		}
		res.Content = string(dataJson)
	} else {
		return fmt.Errorf("验签失败")
	}
	return err
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
	// 计算req.BizContent.OutTradeNo长度
	if len(req.BizContent.OutTradeNo) > 18 {
		return fmt.Errorf("订单号长度不能大于18位:vipspt")
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
	timeFormat := "2006-01-02T15:04:05+08:00"
	createdAt, _ := time.Parse(timeFormat, order["created_at"].(string))
	request := requests.NewCommonRequest()
	request.ApiName = "vipspt.online.trade.order.query"
	request.BizContent = map[string]interface{}{
		"out_trade_no": req.BizContent.OutTradeNo, // 商户订单号(商户交易系统中唯一)
		"shopdate":     createdAt.Format("20060102"),
	}
	return srv.request(request, req, res)
}

func (srv *Trade) Refund(ctx context.Context, req *pb.Request, res *pb.Response) (err error) {
	// 配置参数
	refundOrder, err := mxj.NewMapJson([]byte(req.Config["RefundOrder"]))
	if err != nil {
		return err
	}
	timeFormat := "2006-01-02T15:04:05+08:00"
	createdAt, _ := time.Parse(timeFormat, refundOrder["created_at"].(string))
	refundFee, err := strconv.ParseFloat(req.BizContent.RefundFee, 64)
	if err != nil {
		return err
	}
	if req.BizContent.Title == "" {
		req.BizContent.Title = "退款"
	}
	// 配置参数
	request := requests.NewCommonRequest()
	request.ApiName = "vipspt.online.trade.refund"
	request.BizContent = map[string]interface{}{
		"out_trade_no":   req.BizContent.OutTradeNo,
		"shopdate":       createdAt.Format("20060102"),
		"refund_amount":  decimal.NewFromFloat(refundFee).Div(decimal.NewFromFloat(float64(100))),
		"refund_reason":  req.BizContent.Title,
		"out_request_no": req.BizContent.OutRefundNo, // 商户订单号(商户交易系统中唯一)
	}
	return srv.request(request, req, res)
}

func (srv *Trade) RefundQuery(ctx context.Context, req *pb.Request, res *pb.Response) (err error) {
	// 配置参数
	request := requests.NewCommonRequest()
	request.ApiName = "vipspt.online.trade.refund.query"
	request.BizContent = map[string]interface{}{
		"out_trade_no":   req.BizContent.OutTradeNo,  // 商户订单号(商户交易系统中唯一)
		"out_request_no": req.BizContent.OutRefundNo, // 商户订单号(商户交易系统中唯一)
	}
	return srv.request(request, req, res)
}

func (srv *Trade) JsApi(ctx context.Context, req *pb.Request, res *pb.Response) (err error) {
	totalFee, err := strconv.ParseFloat(req.BizContent.TotalFee, 64)
	if err != nil {
		return err
	}
	// 微信Appid
	wechatAppId := req.BizContent.AppId
	if req.BizContent.AppId == "" {
		wechatAppId = req.Config["WechatAppId"]
	}
	req.Config["NotifyUrl"] = srv.NotifyUrl + "?id=" + req.BizContent.Id
	request := requests.NewCommonRequest()
	request.BizContent = map[string]interface{}{
		"out_trade_no":    req.BizContent.OutTradeNo, // 商户订单号(商户交易系统中唯一)
		"shopdate":        time.Now().Format("20060102"),
		"subject":         req.BizContent.Title,
		"total_amount":    decimal.NewFromFloat(totalFee).Div(decimal.NewFromFloat(float64(100))), // 单位为分
		"seller_id":       req.Config["SubMerId"],
		"timeout_express": "10m",
		"business_code":   "00510030",
	}
	switch req.BizContent.Method {
	case "wechat":
		request.ApiName = "vipspt.online.weixin.pay"
		request.BizContent["sub_openid"] = req.BizContent.OpenId
		request.BizContent["appid"] = wechatAppId
	case "alipay":
		request.ApiName = "vipspt.online.alijsapi.pay"
		request.BizContent["buyer_id"] = req.BizContent.OpenId
	case "unionpay":
		request.ApiName = "vipspt.online.alijsapi.pay"
		request.BizContent["userId"] = req.BizContent.OpenId
		request.BizContent["spbill_create_ip"] = req.BizContent.OpenId
		request.BizContent["allow_repeat_pay"] = "N"
		return fmt.Errorf("暂不支持,银联支付:vipspt")
	}
	return srv.request(request, req, res)
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
	// 配置参数
	switch req.BizContent.Method {
	case "wechat":
	default:
		return fmt.Errorf("暂不支持,只支持微信付款码识别:vipspt")
	}
	request := requests.NewCommonRequest()
	request.ApiName = "vipspt.online.wxpayQueryOpenId"
	request.BizContent = map[string]interface{}{
		"usercode":  req.Config["SubMerId"],
		"auth_code": req.BizContent.AuthCode, // 被扫付款码
		"subAppId":  req.BizContent.AppId,
	}
	return srv.request(request, req, res)
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
