package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/civet148/log"
)

type AliyunDNS struct {
	accessKeyId     string
	accessKeySecret string
	endpoint        string
}

type AliyunResponse struct {
	RequestId string `json:"RequestId"`
	RecordId  string `json:"RecordId"`
}

func NewAliyunDNS(accessKeyId, accessKeySecret string) *AliyunDNS {
	return &AliyunDNS{
		accessKeyId:     accessKeyId,
		accessKeySecret: accessKeySecret,
		endpoint:        "https://alidns.aliyuncs.com/",
	}
}

func (a *AliyunDNS) percentEncode(str string) string {
	encoded := url.QueryEscape(str)
	encoded = strings.Replace(encoded, "+", "%20", -1)
	encoded = strings.Replace(encoded, "*", "%2A", -1)
	encoded = strings.Replace(encoded, "%7E", "~", -1)
	return encoded
}

func (a *AliyunDNS) signature(params map[string]string) string {
	// 对参数进行排序
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建查询字符串
	var query []string
	for _, k := range keys {
		query = append(query, a.percentEncode(k)+"="+a.percentEncode(params[k]))
	}
	queryString := strings.Join(query, "&")

	// 构建签名字符串
	stringToSign := "GET&%2F&" + a.percentEncode(queryString)

	// 计算签名
	h := hmac.New(sha1.New, []byte(a.accessKeySecret+"&"))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signature
}

func (a *AliyunDNS) request(params map[string]string) (*AliyunResponse, error) {
	// 公共参数
	commonParams := map[string]string{
		"Format":           "JSON",
		"Version":          "2015-01-09",
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   strconv.FormatInt(time.Now().UnixNano()/1000000, 10),
		"SignatureVersion": "1.0",
		"AccessKeyId":      a.accessKeyId,
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
	}

	// 合并参数
	finalParams := make(map[string]string)
	for k, v := range commonParams {
		finalParams[k] = v
	}
	for k, v := range params {
		finalParams[k] = v
	}

	// 计算签名
	finalParams["Signature"] = a.signature(finalParams)

	// 构建URL
	var query []string
	for k, v := range finalParams {
		query = append(query, k+"="+url.QueryEscape(v))
	}
	requestURL := a.endpoint + "?" + strings.Join(query, "&")

	// 发送请求
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Infof("阿里云DNS响应: %s", string(body))

	var result AliyunResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (a *AliyunDNS) AddLetsencryptRecord(domain, value string) error {
	// 解析域名，确定主域名和子域名
	parts := strings.Split(domain, ".")
	var mainDomain, rr string

	if len(parts) >= 2 {
		// 一般情况：ruixiangyun.com -> 主域名：ruixiangyun.com, RR: _acme-challenge
		mainDomain = strings.Join(parts[len(parts)-2:], ".")
		rr = "_acme-challenge"

		if len(parts) > 2 {
			// 子域名情况：sub.ruixiangyun.com -> 主域名：ruixiangyun.com, RR: _acme-challenge.sub
			subDomain := strings.Join(parts[:len(parts)-2], ".")
			rr = "_acme-challenge." + subDomain
		}
	} else {
		return fmt.Errorf("无效的域名格式: %s", domain)
	}

	params := map[string]string{
		"Action":     "AddDomainRecord",
		"DomainName": mainDomain,
		"RR":         rr,
		"Type":       "TXT",
		"Value":      value,
	}

	log.Infof("添加DNS记录: 主域名=%s, RR=%s, 值=%s", mainDomain, rr, value)

	_, err := a.request(params)
	if err != nil {
		return fmt.Errorf("添加DNS记录失败: %v", err)
	}

	log.Infof("DNS记录添加成功")
	return nil
}

func (a *AliyunDNS) DeleteLetsencryptRecord(domain string) error {
	// 解析域名，确定主域名和子域名
	parts := strings.Split(domain, ".")
	var mainDomain, rr string

	if len(parts) >= 2 {
		mainDomain = strings.Join(parts[len(parts)-2:], ".")
		rr = "_acme-challenge"

		if len(parts) > 2 {
			subDomain := strings.Join(parts[:len(parts)-2], ".")
			rr = "_acme-challenge." + subDomain
		}
	} else {
		return fmt.Errorf("无效的域名格式: %s", domain)
	}

	params := map[string]string{
		"Action":     "DeleteSubDomainRecords",
		"DomainName": mainDomain,
		"RR":         rr,
		"Type":       "TXT",
	}

	log.Infof("删除DNS记录: 主域名=%s, RR=%s", mainDomain, rr)

	_, err := a.request(params)
	if err != nil {
		return fmt.Errorf("删除DNS记录失败: %v", err)
	}

	log.Infof("DNS记录删除成功")
	return nil
}
