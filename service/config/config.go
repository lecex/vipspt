package config

type Config struct {
	Appid         string `json:"appid"`          //分配给开发者的应用ID
	SecretKey     string `json:"secret_key"`     //私钥
	MerchantId    string `json:"merchant_id"`    // 商户号
	EnterpriseReg string `json:"enterprise_reg"` // 商户注册编码
	SignType      string `json:"sign_type"`      //签名类型
	Sign          string `json:"sign"`           //商户请求参数的签名串
	NotifyUrl     string `json:"notify_url"`     //服务器主动通知商户服务器里指定的页面http/https路径。
	BizContent    string `json:"biz_content"`    //业务请求参数的集合，最大长度不限，除公共参数外所有请求参数都必须放在这个参数中传递，具体参照各产品快速接入文档
	Sandbox       bool   `json:"sandbox"`        // 沙盒
}
