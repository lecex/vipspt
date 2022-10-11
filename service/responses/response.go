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
	"strconv"
	"strings"

	"github.com/clbanning/mxj"
	"github.com/shopspring/decimal"

	"github.com/lecex/vipspt/service/config"
	"github.com/lecex/vipspt/service/requests"
	"github.com/lecex/vipspt/service/util"
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

// {"ret":0,"msg":"操作成功","data":{"third_order_id":"20221011111351886981","out_order_id":"513457061273811891","leshua_order_id":"20221011111351138586","amount":"0.01","status":"2","merchant_id":"307989950941205","enterpriseReg":"NKOt4Ygx","dctime":"2022-10-11 11:13:51","pay_way":"WXZF","sSignature":"tHMMrfoNC7d7jxDdQJR+ViFpleaPvcu+e/mi1hGPvzlEHEXu5IeJ1WGFzba0Af24mGxlunkLmnvqcFNi6E6RU/FkQ1gYsCjqlqOyklz7+d0FBBVnzB/DSv8+0Rs3/AXmLHCJ32bh5w0hmbAJ69+NHJV1KWKQuGIBuWb1LChWNZE="},
// GetVerifySignDataMap 获取 GetVerifySignDataMap 校验后数据数据
func (res *CommonResponse) GetVerifySignDataMap() (m mxj.Map, err error) {
	r, err := res.GetHttpContentMap()
	if err != nil {
		return r, err
	}
	return res.GetSignDataMap()
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
	content, err := mxj.NewMapJson([]byte(res.json))
	if err != nil {
		return nil, err
	}
	if res.Request.ApiName == "pay.pay" {
		data = res.handerVipsptTradePay(content)
	}
	if res.Request.ApiName == "pay.query" {
		data = res.handerVipsptTradeQuery(content)
	}
	if res.Request.ApiName == "pay.refund" {
		data = res.handerVipsptTradeRefund(content)
	}
	if res.Request.ApiName == "pay.refundQuery" {
		data = res.handerVipsptTradeRefundQuery(content)
	}

	data["channel"] = "vipspt" //渠道
	data["content"] = content
	return data, err
}

// {"ret":0,"msg":"操作成功",
// "data":{"third_order_id":"20221011111351886981",
// "out_order_id":"513457061273811891",
// "leshua_order_id":"20221011111351138586",
// "amount":"0.01",
// "status":"2",
// "merchant_id":"307989950941205",
// "enterpriseReg":"NKOt4Ygx",
// "dctime":"2022-10-11 11:13:51",
// "pay_way":"WXZF",
// "sSignature":"tHMMrfoNC7d7jxDdQJR+ViFpleaPvcu+e/mi1hGPvzlEHEXu5IeJ1WGFzba0Af24mGxlunkLmnvqcFNi6E6RU/FkQ1gYsCjqlqOyklz7+d0FBBVnzB/DSv8+0Rs3/AXmLHCJ32bh5w0hmbAJ69+NHJV1KWKQuGIBuWb1LChWNZE="},
// handerVipsptTradePay
func (res *CommonResponse) handerVipsptTradePay(content mxj.Map) mxj.Map {
	data := mxj.New()
	data["status"] = "" // 状态
	data["return_msg"] = ""
	if v, ok := content["msg"]; ok {
		data["return_msg"] = v
	}
	if util.InterfaceToString(content["ret"]) == "0" {
		contentData := content["data"].(map[string]interface{})
		data["return_code"] = SUCCESS
		switch contentData["status"] {
		case "0":
			data["status"] = USERPAYING
		case "2":
			data["status"] = SUCCESS
		case "6":
			data["status"] = CLOSED
		case "7":
			data["status"] = WAITING
		case "8":
			data["status"] = CLOSED
		case "11":
			data["status"] = WAITING
		}
		// string 转 float64
		if v, ok := contentData["amount"]; ok {
			if v1, ok := v.(string); ok {
				if v2, err := strconv.ParseFloat(v1, 64); err == nil {
					total_amt := decimal.NewFromFloat(v2).Mul(decimal.NewFromFloat(float64(100))).IntPart()
					data["total_fee"] = total_amt
				}
			}
		}
		data["buyer_pay_fee"] = data["total_fee"]
		data["bank_trade_no"] = contentData["third_order_id"] // 银行订单
		data["out_trade_no"] = contentData["out_order_id"]
		if v, ok := contentData["dctime"]; ok {
			// 字符串替换-为空
			v = strings.Replace(v.(string), "-", "", -1)
			v = strings.Replace(v.(string), " ", "", -1)
			data["time_end"] = strings.Replace(v.(string), ":", "", -1)
		}

	} else {
		data["return_code"] = "FAIL"
		if data["return_msg"] == "授权码检验错误" {
			data["status"] = WAITING
		}
	}
	return data
}

// res {"ret":0,"msg":"操作成功","data":{"third_order_id":"20221011160916675235",
// "out_order_id":"513457061273811892","leshua_order_id":"20221011160916507196",
// "amount":"0.01","status":"2","payMsg":"支付成功","merchant_id":"307989950941205",
// "enterpriseReg":"NKOt4Ygx","dctime":"2022-10-11 16:09:16.0",
// "pay_way":"WXZF","refNum":"160916507196",
// "refundMsg":"","refund_amount":"","old_third_order_id":"",
// "sSignature":"fQuIZJxq8ghh8pAFME+Hdu3MN6poI/yg00lHAMbh9Vyk16xLw9kU0mZcO+dIRWWTZZ5AwKXB3TT0DZIA2aZ5P9zczdci+2cY4w6N7gHdoNzsjklfyYJ8vpeOUnXojN/tvm5Y1Fiy4d+EqKqaXG0VHo+0mNV6pSIPprDGUL5N4Lg="},
// "totalPage":""}

// handerVipsptTradeQuery
func (res *CommonResponse) handerVipsptTradeQuery(content mxj.Map) mxj.Map {
	data := mxj.New()
	data["status"] = "" // 状态
	data["return_msg"] = ""
	if v, ok := content["msg"]; ok {
		data["return_msg"] = v
	}
	if util.InterfaceToString(content["ret"]) == "0" {
		contentData := content["data"].(map[string]interface{})
		data["return_code"] = SUCCESS
		if v, ok := contentData["payMsg"]; ok {
			data["return_msg"] = v
		}
		switch contentData["status"] {
		case "0":
			data["status"] = USERPAYING
		case "2":
			data["status"] = SUCCESS
		case "6":
			data["status"] = CLOSED
		case "7":
			data["status"] = WAITING
		case "8":
			data["status"] = CLOSED
		case "11":
			data["status"] = WAITING
		}
		// string 转 float64
		if v, ok := contentData["amount"]; ok {
			if v1, ok := v.(string); ok {
				if v2, err := strconv.ParseFloat(v1, 64); err == nil {
					total_amt := decimal.NewFromFloat(v2).Mul(decimal.NewFromFloat(float64(100))).IntPart()
					data["total_fee"] = total_amt
				}
			}
		}
		data["buyer_pay_fee"] = data["total_fee"]
		data["bank_trade_no"] = contentData["third_order_id"] // 银行订单
		data["out_trade_no"] = contentData["out_order_id"]
		if v, ok := contentData["dctime"]; ok {
			// 字符串替换-为空
			v = strings.Replace(v.(string), "-", "", -1)
			v = strings.Replace(v.(string), " ", "", -1)
			data["time_end"] = strings.Replace(v.(string), ":", "", -1)
		}

	} else {
		data["return_code"] = "FAIL"
	}
	return data
}

// {"ret":0,"msg":"操作成功","data":{"third_order_id":"20221011162901020790",
// "dctime":"2022-10-11 16:29:02","refNum":"111351138586",
// "enterpriseReg":"NKOt4Ygx","refundMsg":"退款",
// "out_order_id":"113457061273811892","
// sSignature":"wLglfwasmTRPaOSjSq37tgbkjPdoizMnFR4256WaoCQNa+0o8uflEsIazZ93zdgt2kC8NIL3omfO6QeBcciIL7enb74D1oRuIxjV0ChWL86YS6EUpjQ30lwNRo2ttt06q4dNtyJhPz++I5Thv6mkcfLA22LApd+s4Cob+8GGnGU=",
// "refund_amount":"-0.01",
// "merchant_id":"307989950941205",
// "old_third_order_id":"","pay_way":"WXZF",
// "status":"11"},"totalPage":""}
// handerVipsptTradeRefund
func (res *CommonResponse) handerVipsptTradeRefund(content mxj.Map) mxj.Map {
	data := mxj.New()
	data["status"] = "" // 状态
	data["return_msg"] = ""
	if v, ok := content["msg"]; ok {
		data["return_msg"] = v
	}
	if util.InterfaceToString(content["ret"]) == "0" {
		contentData := content["data"].(map[string]interface{})
		data["return_code"] = SUCCESS
		switch contentData["status"] {
		case "0":
			data["status"] = USERPAYING
		case "2":
			data["status"] = SUCCESS
		case "6":
			data["status"] = CLOSED
		case "7":
			data["status"] = WAITING
		case "8":
			data["status"] = CLOSED
		case "11":
			data["status"] = WAITING
		}
		data["bank_trade_no"] = contentData["third_order_id"] // 银行订单
		data["out_trade_no"] = contentData["out_order_id"]
		if v, ok := contentData["dctime"]; ok {
			// 字符串替换-为空
			v = strings.Replace(v.(string), "-", "", -1)
			v = strings.Replace(v.(string), " ", "", -1)
			data["time_end"] = strings.Replace(v.(string), ":", "", -1)
		}

	} else {
		data["return_code"] = "FAIL"
		if data["return_msg"] == "退款失败：交易金额不符" {
			data["status"] = WAITING
		}
	}
	return data
}

// {"ret":0,"msg":"操作成功","data":{
// "third_order_id":"20221011162901020790",
// "out_order_id":"113457061273811892",
// "leshua_order_id":"20221011111351138586",
// "amount":"-0.01","status":"2","payMsg":"支付成功",
// "merchant_id":"307989950941205","enterpriseReg":"NKOt4Ygx",
// "dctime":"2022-10-11 16:29:02.0","pay_way":"WXZF",
// "refNum":"111351138586","refundMsg":"退款",
// "refund_amount":"-0.01","old_third_order_id":"20221011111351886981",
// "sSignature":"j39jUrt8whxKW+gK7j98euUbTzQ+8FkUP/54Yhn92oGMVvcXqz9G5Qdmrww3jojzpGv03gnDjlG4o79QSXXlxsLIxQtTPNhFa9I1pKkVgt88Sq6P1ofrGIV2bJDJgi1/i+x8bUsdHggbUTq+GMI3Scfe5yBAVQXClW1ghJxzyJQ="},"totalPage":""}

// handerVipsptTradeRefundQuery
func (res *CommonResponse) handerVipsptTradeRefundQuery(content mxj.Map) mxj.Map {
	data := mxj.New()
	data["status"] = "" // 状态
	data["return_msg"] = ""
	if v, ok := content["msg"]; ok {
		data["return_msg"] = v
	}
	if util.InterfaceToString(content["ret"]) == "0" {
		contentData := content["data"].(map[string]interface{})
		data["return_code"] = SUCCESS
		if v, ok := contentData["payMsg"]; ok {
			data["return_msg"] = v
		}
		switch contentData["status"] {
		case "0":
			data["status"] = USERPAYING
		case "2":
			data["status"] = SUCCESS
		case "6":
			data["status"] = CLOSED
		case "7":
			data["status"] = WAITING
		case "8":
			data["status"] = CLOSED
		case "11":
			data["status"] = WAITING
		}
		data["bank_trade_no"] = contentData["third_order_id"] // 银行订单
		data["out_trade_no"] = contentData["out_order_id"]
		if v, ok := contentData["dctime"]; ok {
			// 字符串替换-为空
			v = strings.Replace(v.(string), "-", "", -1)
			v = strings.Replace(v.(string), " ", "", -1)
			data["time_end"] = strings.Replace(v.(string), ":", "", -1)
		}

	} else {
		data["return_code"] = "FAIL"
	}
	return data
}
