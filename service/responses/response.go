/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless resuired by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package responses

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/clbanning/mxj"
	"github.com/micro/go-micro/v2/util/log"
	"github.com/shopspring/decimal"

	"github.com/lecex/vipspt/service/config"
	"github.com/lecex/vipspt/service/requests"
)

const (
	CLOSED     = "CLOSED"     // -1 订单关闭
	USERPAYING = "USERPAYING" // 0	订单支付中
	SUCCESS    = "SUCCESS"    // 1	订单支付成功
	WAITING    = "WAITING"    // 2	系统执行中请等待
)

// CommonResponse 公共回应
type CommonResponse struct {
	Config      *config.Config
	Request     *requests.CommonRequest
	httpContent []byte
	json        string
}

type Map *mxj.Map

// NewCommonResponse 创建新的请求返回
func NewCommonResponse(config *config.Config, request *requests.CommonRequest) (response *CommonResponse) {
	c := &CommonResponse{}
	c.Config = config
	c.Request = request
	return c
}

// GetHttpContentJson 获取 JSON 数据
func (res *CommonResponse) GetHttpContentJson() string {
	return res.json
}

// GetHttpContentMap 获取 MAP 数据
func (res *CommonResponse) GetHttpContentMap() (mxj.Map, error) {
	return mxj.NewMapJson([]byte(res.json))
}

// GetVerifySignDataMap 获取 GetVerifySignDataMap 校验后数据数据
func (res *CommonResponse) GetVerifySignDataMap() (m mxj.Map, err error) {
	r, err := res.GetHttpContentMap()
	if err != nil {
		return r, err
	}
	fmt.Println(122121, r)
	if r["sign"] != nil {
		fmt.Println("debug")
		// ok, err := util.VerifySign(res.GetSignData(), r["sign"].(string), "", res.Config.VipsptPublicKeyData, "RSA")
		// if err != nil {
		// 	return r, err
		// }
		// if ok {
		// 	return res.GetSignDataMap()
		// }
	} else {
		return r, errors.New("res sign is not")
	}
	return
}

// GetSignData 获取 SignData 数据
func (res *CommonResponse) GetSignData() string {
	indexStart := strings.Index(res.json, `response":`)
	indexEnd := strings.Index(res.json, `}}`) + 1
	signData := res.json[indexStart+10 : indexEnd]
	return signData
}

// GetSign 获取 Sign 数据
func (res *CommonResponse) GetSign() (string, error) {
	mv, err := res.GetHttpContentMap()
	if err != nil {
		return "", err
	}
	if _, ok := mv["sign"]; ok { //去掉 xml 外层
		return mv["sign"].(string), err
	}
	return "", err
}

// SetHttpContent 设置请求信息
func (res *CommonResponse) SetHttpContent(httpContent []byte, dataType string) {
	res.httpContent = httpContent
	switch dataType {
	case "xml":
		mv, _ := mxj.NewMapXml(res.httpContent) // unmarshal
		var str interface{}
		if _, ok := mv["xml"]; ok { //去掉 xml 外层
			str = mv["xml"]
		} else {
			str = mv
		}
		jsonStr, _ := json.Marshal(str)
		res.json = string(jsonStr)
	case "string":
		res.json = string(res.httpContent)
	}
}

// GetSignDataMap 获取 MAP 数据
func (res *CommonResponse) GetSignDataMap() (mxj.Map, error) {
	data := mxj.New()
	content, err := mxj.NewMapJson([]byte(res.GetSignData()))
	if err != nil {
		return nil, err
	}
	if res.Request.ApiName == "vipspt.online.barcodepay" {
		data = res.handerVipsptTradePay(content)
	}
	if res.Request.ApiName == "vipspt.online.trade.order.query" {
		data = res.handerVipsptTradeQuery(content)
	}
	if res.Request.ApiName == "vipspt.online.trade.refund" {
		data = res.handerVipsptTradeRefund(content)
	}
	if res.Request.ApiName == "vipspt.online.trade.refund.query" {
		data = res.handerVipsptTradeRefundQuery(content)
	}
	if res.Request.ApiName == "pay.qrcode" {
		data = res.handerVipsptTradeQrcode(content)
	}
	if res.Request.ApiName == "pay.openid" {
		data = res.handerVipsptTradeOpenid(content)
	}
	if res.Request.ApiName == "vipspt.online.weixin.pay" || res.Request.ApiName == "vipspt.online.alijsapi.pay" {
		data = res.handerVipsptTradeJsApi(content)
	}

	data["channel"] = "vipspt" //渠道
	data["content"] = content
	return data, err
}

// data{
// 	channel			//	通道内容		alipay、wechat、vipspt
// 	content			//	第三方返回内容 	{}
// 	return_code		//	返回代码 		SUCCESS
// 	return_msg		//	返回消息		支付失败
// 	status			//	下单状态 		【SUCCESS成功、CLOSED关闭、USERPAYING等待用户付款、WAITING系统繁忙稍后查询】
// 	total_fee		//  订单金额		88
// 	refund_fee 		//  退款金额		10
//  buyer_pay_fee// 用户实际付款金额  66
// 	trade_no 		// 	渠道交易编号 	2013112011001004330000121536
// 	out_trade_no	// 	商户订单号		T1024501231476
//  out_refund_no	//  商户退款单号	T1024501231476_T
// 	wechat_open_id		//  微信openid		[oUpF8uN95-Ptaags6E_roPHg7AG
//  wechat_is_subscribe 	//  微信是否微信关注公众号
// 	alipay_logon_id  //	支付宝账号		158****1562
//  alipay_user_id  //	买家在支付宝的用户id	2088101117955611
// 	time_end		//  支付完成时间	20141030133525
//  wechat_package // 微信支付包
// }
// handerVipsptTradePay
func (res *CommonResponse) handerVipsptTradePay(content mxj.Map) mxj.Map {
	data := mxj.New()
	data["status"] = "" // 状态
	data["return_msg"] = ""
	if v, ok := content["msg"]; ok {
		data["return_msg"] = v
	}
	if v, ok := content["sub_msg"]; ok {
		data["return_msg"] = v
	}
	if content["code"] == "10000" {
		data["return_code"] = SUCCESS
		switch content["trade_status"] {
		case "WAIT_BUYER_PAY":
			data["status"] = USERPAYING
		case "TRADE_CLOSED":
			data["status"] = CLOSED
		case "TRADE_SUCCESS":
			data["status"] = SUCCESS
		case "TRADE_PART_REFUND":
			data["status"] = SUCCESS
		case "TRADE_ALL_REFUND":
			data["status"] = SUCCESS
		case "TRADE_PROCESS":
			data["status"] = WAITING
		case "TRADE_FAILD":
			data["status"] = CLOSED
		}
		if v, ok := content["trade_status_ext"]; ok {
			if v == "TRADE_USERPAYING" {
				data["status"] = USERPAYING
			}
		}
		// string 转 float64
		if v, ok := content["total_amount"]; ok {
			if v1, ok := v.(string); ok {
				if v2, err := strconv.ParseFloat(v1, 64); err == nil {
					total_amt := decimal.NewFromFloat(v2).Mul(decimal.NewFromFloat(float64(100))).IntPart()
					data["total_fee"] = total_amt
				}
			}
		}
		if v, ok := content["settlement_amount"]; ok {
			if v1, ok := v.(string); ok {
				if v2, err := strconv.ParseFloat(v1, 64); err == nil {
					i := decimal.NewFromFloat(v2).Mul(decimal.NewFromFloat(float64(100))).IntPart()
					data["buyer_pay_fee"] = i
				}
			}
		} else {
			data["buyer_pay_fee"] = data["total_fee"]
		}
		data["bank_trade_no"] = content["trade_no"] // 银行订单
		data["out_trade_no"] = content["out_trade_no"]
		if v, ok := content["account_date"]; ok {
			// 字符串替换-为空
			data["time_end"] = strings.Replace(v.(string), "-", "", -1)
		}
		if v, ok := content["openid"]; ok {
			data["wechat_open_id"] = v
		}

	} else {
		data["return_code"] = "FAIL"
		if content["sub_code"] == "3161" {
			data["status"] = WAITING
		}
		if content["sub_code"] == "3155" {
			data["status"] = WAITING
		}
		if content["sub_code"] == "3172" {
			data["status"] = WAITING
		}

	}
	return data
}

// handerVipsptTradeQuery
func (res *CommonResponse) handerVipsptTradeQuery(content mxj.Map) mxj.Map {
	data := mxj.New()
	data["status"] = "" // 状态
	data["return_msg"] = ""
	if v, ok := content["msg"]; ok {
		data["return_msg"] = v
	}
	if v, ok := content["sub_msg"]; ok {
		data["return_msg"] = v
	}
	if content["code"] == "10000" {
		data["return_code"] = SUCCESS
		switch content["trade_status"] {
		case "WAIT_BUYER_PAY":
			data["status"] = USERPAYING
		case "TRADE_CLOSED":
			data["status"] = CLOSED
		case "TRADE_SUCCESS":
			data["status"] = SUCCESS
		case "TRADE_PART_REFUND":
			data["status"] = SUCCESS
		case "TRADE_ALL_REFUND":
			data["status"] = SUCCESS
		case "TRADE_PROCESS":
			data["status"] = WAITING
		case "TRADE_FAILD":
			data["status"] = CLOSED
		}
		if v, ok := content["trade_status_ext"]; ok {
			if v == "TRADE_USERPAYING" {
				data["status"] = USERPAYING
			}
		}
		// string 转 float64
		if v, ok := content["total_amount"]; ok {
			if v1, ok := v.(string); ok {
				if v2, err := strconv.ParseFloat(v1, 64); err == nil {
					total_amt := decimal.NewFromFloat(v2).Mul(decimal.NewFromFloat(float64(100))).IntPart()
					data["total_fee"] = total_amt
				}
			}
		}
		if v, ok := content["settlement_amount"]; ok {
			if v1, ok := v.(string); ok {
				if v2, err := strconv.ParseFloat(v1, 64); err == nil {
					i := decimal.NewFromFloat(v2).Mul(decimal.NewFromFloat(float64(100))).IntPart()
					data["buyer_pay_fee"] = i
				}
			}
		} else {
			data["buyer_pay_fee"] = data["total_fee"]
		}
		data["bank_trade_no"] = content["trade_no"] // 银行订单
		data["out_trade_no"] = content["out_trade_no"]
		if v, ok := content["account_date"]; ok {
			// 字符串替换-为空
			data["time_end"] = strings.Replace(v.(string), "-", "", -1)
		}
		if v, ok := content["openid"]; ok {
			data["wechat_open_id"] = v
		}

	} else {
		data["return_code"] = "FAIL"
		if content["sub_code"] == "3161" {
			data["status"] = WAITING
		}
		if content["sub_code"] == "3155" {
			data["status"] = WAITING
		}
		if content["sub_code"] == "3172" {
			data["status"] = WAITING
		}
		if content["sub_code"] == "ACQ.QUERY_NO_RESULT" {
			data["status"] = CLOSED
		}

	}
	return data
}

// {"appid":"00000051",
// "chnltrxid":"2022071722001476811402811792",
// "cusid":"990581007426001",
// "errmsg":"金额错误,当前可退货金额为:0.00",
// "fintime":"20220718112948",
// "randomstr":"414084924970",
// "reqsn":"51345706127381181882-1",
// "retcode":"SUCCESS","sign":"=",
// "trxcode":"VSP513",
// "trxstatus":"3999"
// }
// handerVipsptTradeRefund
func (res *CommonResponse) handerVipsptTradeRefund(content mxj.Map) mxj.Map {
	data := mxj.New()
	data["status"] = "WAITING" // 状态
	data["return_msg"] = ""
	if v, ok := content["msg"]; ok {
		data["return_msg"] = v
	}
	if v, ok := content["sub_msg"]; ok {
		data["return_msg"] = v
	}
	if content["code"] == "10000" {
		data["return_code"] = SUCCESS
		data["status"] = WAITING
		// string 转 float64
		if v, ok := content["refund_amount"]; ok {
			if v1, ok := v.(string); ok {
				if v2, err := strconv.ParseFloat(v1, 64); err == nil {
					i := decimal.NewFromFloat(v2).Mul(decimal.NewFromFloat(float64(100))).IntPart()
					data["refund_fee"] = i
				}
			}
		}
		data["bank_trade_no"] = content["refundsn"] // 银行订单
		data["out_refund_no"] = content["out_request_no"]
		data["out_trade_no"] = content["out_trade_no"]
		if v, ok := content["account_date"]; ok {
			// 字符串替换-为空
			data["time_end"] = strings.Replace(v.(string), "-", "", -1)
		}

	} else {
		data["return_code"] = "FAIL"
	}
	return data
}

// {"code":"10000","msg":"Success",
// "trade_no":"02Q220909725664900","out_trade_no":"51345706127381181888",
// "out_request_no":"51345706127381181888-1","refund_state":"success",
// "account_date":"2022-09-09","refund_reason":"退款",
// "total_amount":0.01,"refund_amount":0.01,"funds_state":"success",
// "funds_dynamics":"[{\"channelRecvSn\":\"50100703022022090924690384069\",\
// "channelRecvTime\":1662694628000,\"channelSendSn\":\"Q10222090972903954K1\",\
// "marketingRefundDetail\":\"{\\\"fee_type\\\":\\\"CNY\\\",\\\"refund_fee\\\":0.01,\\\
// "cash_refund_fee\\\":0.01,\\\"settlement_refund_fee\\\":0.01,\\\
// "coupon_refund_fee\\\":0.0}\",\"refundamount\":0.01,\
// "refundsn\":\"02R22090972903954K\",\"sendChannelTime\":1662694620000,\
// "state\":\"00\"}]","src_fee_flag":"01","has_refund_src_fee":0.0,
// "payee_fee_flag":"01","has_refund_payee_fee":0.0,"payer_fee_flag":"01",
// "has_refund_payer_fee":0.0,"markting_refund_detail":"{\"fee_type\":\"CNY\",\
// "refund_fee\":0.01,\"cash_refund_fee\":0.01,\"settlement_refund_fee\":0.01,\
// "coupon_refund_fee\":0.0}",
// "channel_recv_time":"2022-09-09 11:37:08"}

// handerVipsptTradeRefundQuery
func (res *CommonResponse) handerVipsptTradeRefundQuery(content mxj.Map) mxj.Map {
	data := mxj.New()
	data["status"] = "WAITING" // 状态
	data["return_msg"] = ""
	if v, ok := content["msg"]; ok {
		data["return_msg"] = v
	}
	if v, ok := content["sub_msg"]; ok {
		data["return_msg"] = v
	}
	if content["code"] == "10000" {
		data["return_code"] = SUCCESS
		switch content["refund_state"] {
		case "in_process":
			data["status"] = WAITING
		case "success":
			data["status"] = SUCCESS
		case "fail_due_manual_close":
			data["status"] = WAITING
		case "fail_to_manual_deal":
			data["status"] = CLOSED
		case "fail":
			data["status"] = CLOSED
		}
		// string 转 float64
		if v, ok := content["refund_amount"]; ok {
			if v1, ok := v.(string); ok {
				if v2, err := strconv.ParseFloat(v1, 64); err == nil {
					i := decimal.NewFromFloat(v2).Mul(decimal.NewFromFloat(float64(100))).IntPart()
					data["refund_fee"] = i
				}
			}
		}
		// data["bank_trade_no"] = content["refundsn"] // 银行订单
		data["out_refund_no"] = content["out_request_no"]
		data["out_trade_no"] = content["out_trade_no"]
		if v, ok := content["account_date"]; ok {
			// 字符串替换-为空
			data["time_end"] = strings.Replace(v.(string), "-", "", -1)
		}

	} else {
		data["return_code"] = "FAIL"
	}
	return data
}

func (res *CommonResponse) handerVipsptTradeQrcode(content mxj.Map) mxj.Map {
	data := mxj.New()
	data["return_msg"] = ""
	if v, ok := content["msg"]; ok {
		data["return_msg"] = v
	}
	if v, ok := content["sub_msg"]; ok {
		data["return_msg"] = v
	}
	if content["return_code"] == "0" {
		data["return_code"] = SUCCESS
		data["qr_code"] = content["qrcode"]
	} else {
		data["return_code"] = "FAIL"
		if res.InterfaceToString(content["return_code"]) == "400019" {
			data["status"] = CLOSED
		}
	}
	return data
}

// handerVipsptTradeOpenid
func (res *CommonResponse) handerVipsptTradeOpenid(content mxj.Map) mxj.Map {
	data := mxj.New()
	data["return_msg"] = ""
	if v, ok := content["msg"]; ok {
		data["return_msg"] = v
	}
	if v, ok := content["sub_msg"]; ok {
		data["return_msg"] = v
	}
	if content["retcode"] == "SUCCESS" {
		data["return_code"] = SUCCESS
		data["wechat_open_id"] = content["acct"]
	} else {
		data["return_code"] = "FAIL"
	}
	return data
}

// ParseNotifyResult 解析异步通知
func (res *CommonResponse) InterfaceToString(v interface{}) string {
	switch v.(type) {
	case string:
		return v.(string)
	case int:
		return strconv.Itoa(v.(int))
	case int64:
		return strconv.FormatInt(v.(int64), 10)
	case float32:
		return strconv.FormatFloat(v.(float64), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v.(float64), 'f', -1, 64)
	}
	return ""
}

// HanderVipsptNotify
func (res *CommonResponse) HanderVipsptNotify(content mxj.Map) mxj.Map {
	// 查询 交易结果标志：0：支付中请稍后查询，1：支付成功，2：支付失败，3：已撤销，4：撤销中请稍后查询，5：已全额退款，6：已部分退款，7：退款中请稍后查询
	data := mxj.New()
	data["return_msg"] = ""
	data["return_code"] = SUCCESS

	switch content["trade_status"] {
	case "WAIT_BUYER_PAY":
		data["status"] = USERPAYING
	case "TRADE_CLOSED":
		data["status"] = CLOSED
	case "TRADE_SUCCESS":
		data["status"] = SUCCESS
	case "TRADE_PART_REFUND":
		data["status"] = SUCCESS
	case "TRADE_ALL_REFUND":
		data["status"] = SUCCESS
	case "TRADE_PROCESS":
		data["status"] = WAITING
	case "TRADE_FAILD":
		data["status"] = CLOSED
	}
	if v, ok := content["trade_status_ext"]; ok {
		if v == "TRADE_USERPAYING" {
			data["status"] = USERPAYING
		}
	}
	// string 转 float64
	if v, ok := content["total_amount"]; ok {
		if v1, ok := v.(string); ok {
			if v2, err := strconv.ParseFloat(v1, 64); err == nil {
				total_amt := decimal.NewFromFloat(v2).Mul(decimal.NewFromFloat(float64(100))).IntPart()
				data["total_fee"] = total_amt
			}
		}
	}
	if v, ok := content["settlement_amount"]; ok {
		if v1, ok := v.(string); ok {
			if v2, err := strconv.ParseFloat(v1, 64); err == nil {
				i := decimal.NewFromFloat(v2).Mul(decimal.NewFromFloat(float64(100))).IntPart()
				data["buyer_pay_fee"] = i
			}
		}
	} else {
		data["buyer_pay_fee"] = data["total_fee"]
	}
	if v, ok := content["out_trade_no"]; ok {
		data["out_trade_no"] = v
	}
	if v, ok := content["trade_no"]; ok {
		data["bank_trade_no"] = v
	}
	if v, ok := content["channel_recv_sn"]; ok {
		data["trade_no"] = v
	}
	if v, ok := content["account_date"]; ok {
		data["time_end"] = v
	}
	if v, ok := content["openid"]; ok {
		data["wechat_open_id"] = v
	}
	if v, ok := content["buyer_user_id"]; ok {
		data["alipay_user_id"] = v
	}

	data["channel"] = "vipspt" //渠道
	data["content"] = content
	return data
}

// 2022-07-26 17:33:41  level=info [Vipspt[PostForm]res
// {"appid":"00244520",
// "cusid":"6604660739906DX",
// "payinfo":"{\"appId\":\"wxb3fa424b649563b5\",\"timeStamp\":\"1658828021\",
// \"nonceStr\":\"d0f3c544dea74c7aa3e9becda85f223c\",
// \"package\":\"prepay_id=wx261733409734034e23ccf83d002d3f0000\",
// \"signType\":\"RSA\",\"paySign\":\"l9kA65XWVdeTJnJrWXHwSHmCPqALiSVnTVZdQtXJwTMEbr2
// MxframNrR+nDxxuWzX5TeozFOFHIHxWw7z1lgklJdlA2YwQij66IuP5VF5QBxsok3K9PmphgFF4ojtQaEg5gmK
// uDA7pxjDGJDBYg4uPjfRMrLIa/xQoX2E8y5BbQmCUOENyE5S/ExHBr1sh6XGLz2FZSZb1Ns0lTPkKMjYvGtESr
// K6NsHU8kGQC7o1fFmqX4prCEXo9b+0wCyVyJ43/l+E3BF/k4I2DHdgYDl9o+O5hx/AuUI11U011nOsIZBOFYXs
// rD4SglsXLma4rj0coBwi9fYk36skyYzTrQibA==\"}",
// "randomstr":"483178930448",
// "reqsn":"20220726173340055954",
// "retcode":"SUCCESS",
// "sign":"izJ1z0UsV8u/WOs/bTaL0g53ueDzeyq4V71w0Ebhn7QrtWR7GK
// 6nB72trTwCOPDXnhZdixbAw7181jAvydzQdZ6LHYbO7N8MyUsgMz9knpIffA0InX57LW3mIa6b
// dQ6XZ/B9EEV0p4iRy9LBjoDGrSb+3Yz/zI15sQmBX6ZUDC4=",
// "trxid":"220726112833402007",
// "trxstatus":"0000"} <nil>]
// 2022-07-27 09:24:12  level=info [Vipspt[PostForm] https://vsp.vipspt.com/apiweb/unitorder/pay
// map[acct:okCtS6IyyODgL6EyAI3HQLUEN-cs
// appid:00244520 body:C2B二维码支付 cusid:6604660739906DX
// paytype:W02 randomstr:1f7ffb246fb1478e8449a5fc526a43b8
// reqsn:2022072709241216779 sign:UysM+7+BT9HUnIUBJfw7Gr1MByft+tsepot0/5A8E5vfT
// BsNpoteSvSqejey0D3q5e9IIQvSMfLWlsro8ihNsD5ZNfVsj8o083/6WrHs9ZPgzSyLY/j7gH+qhApb6cwpFORmTUFciv
// KaSsyOdt+Dcsc23P89fo8BNNYXHipSDyLE1Tp02duXXF8nzsiAX37wxzNLWOlfEZrF2BYjLkf9LHxsLrUZ1wIWHYhjH2So5bflju/nW0
// cc6C3575le10xlvtSiF3nmQoKpPadJDR4FV2Dnx9NVGTm7eTXepXa/iRA2vfZYQugxW8dEOadw06sQRt7G3QkYVw56b6TI2XNy3g==
// signtype:RSA sub_appid:
// trxamt:1 version:11]]

// handerVipsptTradeB2cJsApi
func (res *CommonResponse) handerVipsptTradeJsApi(content mxj.Map) mxj.Map {
	data := mxj.New()
	data["status"] = WAITING // 状态
	data["return_msg"] = ""
	if v, ok := content["msg"]; ok {
		data["return_msg"] = v
	}
	if v, ok := content["sub_msg"]; ok {
		data["return_msg"] = v
	}
	if content["code"] == "10000" {
		data["return_code"] = SUCCESS
		switch content["trade_status"] {
		case "WAIT_BUYER_PAY":
			data["status"] = USERPAYING
		}
		// string 转 float64
		if v, ok := content["total_amount"]; ok {
			if v1, ok := v.(string); ok {
				if v2, err := strconv.ParseFloat(v1, 64); err == nil {
					total_amt := decimal.NewFromFloat(v2).Mul(decimal.NewFromFloat(float64(100))).IntPart()
					data["total_fee"] = total_amt
				}
			}
		}
		data["bank_trade_no"] = content["trade_no"] // 银行订单
		data["out_trade_no"] = content["out_trade_no"]
		if v, ok := content["jsapi_pay_info"]; ok {
			payInfo, err := mxj.NewMapJson([]byte(v.(string)))
			if err != nil {
				log.Error("handerIcbcTradeJftJsApi", err)
			}
			if v, ok := payInfo["tradeNO"]; ok {
				data["prepay_id"] = v
			}
			if _, ok := payInfo["appId"]; ok {
				data["wechat_package"] = v.(string)
			}
		}

	} else {
		data["return_code"] = "FAIL"

	}
	return data
}
