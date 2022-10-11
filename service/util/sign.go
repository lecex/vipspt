package util

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"hash"
	"io/ioutil"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

const (
	SignType_MD5    = "MD5"
	SignType_SHA1   = "SHA1"
	SignType_SHA256 = "SHA256"
)

// VerifySign 验证支付
func VerifySign(signData string, sign string, pubCer string, pubData string, signType string) (ok bool, err error) {
	var (
		h     hash.Hash
		hashs crypto.Hash
		// block     *pem.Block
		// pubKey    interface{}
		// publicKey *rsa.PublicKey
		// keyOk     bool
	)
	signBytes, _ := base64.StdEncoding.DecodeString(sign)
	var certD []byte
	if pubCer != "" {
		certD, err = ioutil.ReadFile(pubCer)
		if err != nil {
			return ok, fmt.Errorf("unable to find cert path=%s, error=%v", pubCer, err)
		}
	}
	if pubData != "" {
		certD, err = base64.StdEncoding.DecodeString(pubData)
		if err != nil {
			return ok, fmt.Errorf("certData 公钥文件转码错误, error=%v", err)
		}
	}
	// block, _ := pem.Decode(certD) // .pem
	_, rest := pem.Decode(certD) // .cer
	var cert *x509.Certificate
	cert, _ = x509.ParseCertificate(rest)
	publicKey := cert.PublicKey.(*rsa.PublicKey)
	switch signType {
	case "RSA":
		hashs = crypto.SHA1
	case "RSA2":
		hashs = crypto.SHA256
	default:
		hashs = crypto.SHA256
	}
	h = hashs.New()
	h.Write([]byte(signData))
	err = rsa.VerifyPKCS1v15(publicKey, hashs, h.Sum(nil), signBytes)
	if err != nil {
		return ok, err
	}
	return true, err
}

// EncodeSignParams 编码符号参数
func EncodeSignParams(params map[string]interface{}) string {
	var buf strings.Builder
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := params[k]
		if v == "" {
			continue
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(InterfaceToString(v))
		buf.WriteByte('&')
	}
	return buf.String()[:buf.Len()-1]
}

// Sign 开发平台签名支付签名.
func Sign(params map[string]interface{}, secretKey string) (sign string, err error) {
	encodeSignParams := EncodeSignParams(params) + secretKey
	sum := sha256.Sum256([]byte(encodeSignParams))
	sign = hex.EncodeToString([]byte(sum[:]))
	// sign大写
	sign = strings.ToUpper(sign)
	return sign, nil
}

// ParseNotifyResult 解析异步通知
func InterfaceToString(v interface{}) string {
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
	case decimal.Decimal:
		return v.(decimal.Decimal).String()
	}
	return ""
}

// FormatPrivateKey 格式化 普通应用秘钥
func FormatPrivateKey(privateKey string) (pKey string) {
	var buffer strings.Builder
	buffer.WriteString("-----BEGIN RSA PRIVATE KEY-----\n")
	rawLen := 64
	keyLen := len(privateKey)
	raws := keyLen / rawLen
	temp := keyLen % rawLen
	if temp > 0 {
		raws++
	}
	start := 0
	end := start + rawLen
	for i := 0; i < raws; i++ {
		if i == raws-1 {
			buffer.WriteString(privateKey[start:])
		} else {
			buffer.WriteString(privateKey[start:end])
		}
		buffer.WriteByte('\n')
		start += rawLen
		end = start + rawLen
	}
	buffer.WriteString("-----END RSA PRIVATE KEY-----\n")
	pKey = buffer.String()
	return
}

// FormatURLParam 格式化请求URL参数
func FormatURLParam(params map[string]interface{}) (urlParam string) {
	v := url.Values{}
	for key, value := range params {
		v.Add(key, InterfaceToString(value))
	}
	return v.Encode()
}

// EncodeSignParams 编码符号参数
func EncodeSignParamsNotEmpty(params map[string]interface{}) string {
	var buf strings.Builder
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := params[k]
		// if v == "" {
		// 	continue
		// }
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(InterfaceToString(v))
		buf.WriteByte('&')
	}
	return buf.String()[:buf.Len()-1]
}
