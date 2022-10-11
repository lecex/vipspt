package common

import (
	"fmt"
	"time"

	"github.com/lecex/vipspt/service/config"
	"github.com/lecex/vipspt/service/requests"
	"github.com/lecex/vipspt/service/responses"
	"github.com/lecex/vipspt/service/util"
	"github.com/micro/go-micro/v2/util/log"
)

var apiUrlsMch = map[string]string{
	"pay.pay":         "/payOpen/bToC",      //付款码支付
	"pay.query":       "/payOpen/query.do",  //统一查询接口
	"pay.refund":      "/payOpen/refund.do", //统一退款接口
	"pay.refundQuery": "/payOpen/query.do",  //统一退款查询接口
}

// Common 公共封装
type Common struct {
	Config   *config.Config
	Requests *requests.CommonRequest
}

// Action 创建新的公共连接
func (c *Common) Action(response *responses.CommonResponse) (err error) {
	return c.Request(response)
}

// APIBaseURL 默认 API 网关
func (c *Common) APIBaseURL() string {
	con := c.Config
	if con.Sandbox { // 沙盒模式
		return "http://47.107.41.218:8093"
	}
	return "http://www.vipspt.cn"
}

// ApiUrl 创建 ApiUrl
func (c *Common) ApiUrl() (apiUrl string, err error) {
	req := c.Requests
	if u, ok := apiUrlsMch[req.ApiName]; ok {
		apiUrl = c.APIBaseURL() + u
	} else {
		err = fmt.Errorf("ApiName 不存在请检查。")
	}
	return
}

// Request 执行请求
// AppId        string `json:"app_id"`         //工行开发平台分配给开发者的应用ID
// Method       string `json:"method"`         //接口名称
// Format       string `json:"format"`         //仅支持 JSON
// Charset      string `json:"charset"`        //请求使用的编码格式，如utf-8,gbk,gb2312等，推荐使用 utf-8
// SignType     string `json:"sign_type"`      //商户生成签名字符串所使用的签名算法类型，目前支持RSA2和RSA，推荐使用 RSA2
// Sign         string `json:"sign"`           //商户请求参数的签名串
// Timestamp    string `json:"timestamp"`      //发送请求的时间，格式"yyyy-MM-dd HH:mm:ss"
// Version      string `json:"version"`        //调用的接口版本，固定为：1.0
// NotifyUrl    string `json:"notify_url"`     //工行开发平台服务器主动通知商户服务器里指定的页面http/https路径。
// BizContent   string `json:"biz_content"`    //业务请求参数的集合，最大长度不限，除公共参数外所有请求参数都必须放在这个参数中传递，具体参照各产品快速接入文档
// ReturnUrl    string `json:"return_url"`     //HTTP/HTTPS开头字符串
func (c *Common) Request(response *responses.CommonResponse) (err error) {
	con := c.Config
	req := c.Requests
	apiUrl, err := c.ApiUrl()
	if err != nil {
		return err
	}
	sign, err := util.Sign(req.BizContent, con.SecretKey) // 开发签名
	if err != nil {
		return err
	}
	// 构建配置参数
	params := map[string]interface{}{
		"appid": con.Appid,
		// "appsecret": con.SecretKey,
		"timeStamp": time.Now().UnixNano() / 1e6, // 毫米时间戳
		"sign":      sign,
		"data":      req.BizContent,
	}
	log.Info("Vipspt[PostJSON]", apiUrl, params)
	res, err := util.PostJSON(apiUrl, params)
	log.Info("Vipspt[PostJSON]res", string(res), err)
	if err != nil {
		return err
	}
	response.SetHttpContent(res, "string")
	return
}
