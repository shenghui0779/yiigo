package yiigo

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"models/jsons"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

/**
 * 爬虫基础类 [包括：http、https(CA证书)、cookie、验证码处理]
 * 做爬虫时需用到另外两个库：
 * 1、jbk 转 utf8：gopkg.in/iconv.v1 [https://github.com/qiniu/iconv]
 * 2、页面 dom 处理：github.com/PuerkitoBio/goquery
 * CAPath string CA证书存放路径 [默认 certificate 目录，证书需用 openssl 转化为 pem格式]
 * CookiePath string cookie存放路径 [默认 cookies 目录]
 * 验证码图片默认存放路径为 verifycode 目录
 */
type SpiderBase struct {
	CAPath     string
	CookiePath string
}

/**
 * @title get请求
 * @param {string} httpUrl [请求地址]
 * @param {string} host [请求头部Host]
 * @param {bool} setCookie [请求是否需要加cookie]
 * @param {bool} saveCookie [是否保存返回的cookie]
 * @param {bool} clearOldCookie [是否需要清空原来的cookie]
 * @param {string} referer [请求头部referer]
 * @return io.ReadCloser
 */
func (this *SpiderBase) HttpGet(httpUrl string, host string, setCookie bool, saveCookie bool, clearOldCookie bool, referer ...string) (io.ReadCloser, error) {
	req, httpErr := http.NewRequest("GET", httpUrl, nil)

	if httpErr != nil {
		LogError("new http get error: ", httpErr.Error())
		return nil, httpErr
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Connection", "Keep-Alive")
	req.Header.Set("Host", host)

	if len(referer) > 0 {
		req.Header.Set("Referer", referer[0])
	}

	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0)")

	if setCookie {
		this.SetHttpCookie(req)
	}

	//忽略对服务端传过来的数字证书进行校验
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   20 * time.Second,
	}
	res, clientDoErr := client.Do(req)

	if clientDoErr != nil {
		LogError("client do http get error: ", clientDoErr.Error())
		return nil, clientDoErr
	}

	if saveCookie {
		this.SaveHttpCookie(res.Cookies(), clearOldCookie)
	}

	return res.Body, nil
}

/**
 * @title post请求
 * @param {string} httpUrl [请求地址]
 * @param {string} host [请求头部Host]
 * @param {url.Values} v [post参数]
 * @param {bool} setCookie [请求是否需要加cookie]
 * @param {bool} saveCookie [是否保存返回的cookie]
 * @param {bool} clearOldCookie [是否需要清空原来的cookie]
 * @param {string} referer [请求头部referer]
 * @return io.ReadCloser
 */
func (this *SpiderBase) HttpPost(httpUrl string, host string, v url.Values, setCookie bool, saveCookie bool, clearOldCookie bool, referer ...string) (io.ReadCloser, error) {
	postParam := strings.NewReader(v.Encode())
	req, httpErr := http.NewRequest("POST", httpUrl, postParam)

	if httpErr != nil {
		LogError("new http post error: ", httpErr.Error())
		return nil, httpErr
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Connection", "Keep-Alive")
	req.Header.Set("Host", host)

	if len(referer) > 0 {
		req.Header.Set("Referer", referer[0])
	}

	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0)")

	if setCookie {
		this.SetHttpCookie(req)
	}

	//忽略对服务端传过来的数字证书进行校验
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   20 * time.Second,
	}
	res, clientDoErr := client.Do(req)

	if clientDoErr != nil {
		LogError("client do post error: ", clientDoErr.Error())
		return nil, clientDoErr
	}

	if saveCookie {
		this.SaveHttpCookie(res.Cookies(), clearOldCookie)
	}

	return res.Body, nil
}

/**
 * @title get请求 [https 需要CA证书，用openssl转换成pem格式：cert.pem、key.pem]
 * @param {string} httpUrl [请求地址]
 * @param {string} host [请求头部Host]
 * @param {bool} setCookie [请求是否需要加cookie]
 * @param {bool} saveCookie [是否保存返回的cookie]
 * @param {bool} clearOldCookie [是否需要清空原来的cookie]
 * @param {string} referer [请求头部referer]
 * @return io.ReadCloser
 */
func (this *SpiderBase) HttpsGet(httpUrl string, host string, setCookie bool, saveCookie bool, clearOldCookie bool, referer ...string) (io.ReadCloser, error) {
	req, httpErr := http.NewRequest("GET", httpUrl, nil)

	if httpErr != nil {
		LogError("new http get error: ", httpErr.Error())
		return nil, httpErr
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Connection", "Keep-Alive")
	req.Header.Set("Host", host)

	if len(referer) > 0 {
		req.Header.Set("Referer", referer[0])
	}

	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0)")

	if setCookie {
		this.SetHttpCookie(req)
	}

	certFile, _ := filepath.Abs(fmt.Sprintf("certificate/%s/cert.pem", this.CAPath))
	keyFile, _ := filepath.Abs(fmt.Sprintf("certificate/%s/key.unencrypted.pem", this.CAPath))

	cert, certErr := tls.LoadX509KeyPair(certFile, keyFile)

	if certErr != nil {
		LogError("load x509 keypair error: ", certErr.Error())
		return nil, certErr
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
		},
		DisableCompression: true,
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   20 * time.Second,
	}
	res, clientDoErr := client.Do(req)

	if clientDoErr != nil {
		LogError("client do http get error: ", clientDoErr.Error())
		return nil, clientDoErr
	}

	if saveCookie {
		this.SaveHttpCookie(res.Cookies(), clearOldCookie)
	}

	return res.Body, nil
}

/**
 * @title post请求 [https 需要CA证书，用openssl转换成pem格式：cert.pem、key.pem]
 * @param {string} httpUrl [请求地址]
 * @param {string} host [请求头部Host]
 * @param {url.Values} v [post参数]
 * @param {bool} setCookie [请求是否需要加cookie]
 * @param {bool} saveCookie [是否保存返回的cookie]
 * @param {bool} clearOldCookie [是否需要清空原来的cookie]
 * @param {string} referer [请求头部referer]
 * @return io.ReadCloser
 */
func (this *SpiderBase) HttpsPost(httpUrl string, host string, v url.Values, setCookie bool, saveCookie bool, clearOldCookie bool, referer ...string) (io.ReadCloser, error) {
	postParam := strings.NewReader(v.Encode())
	req, httpErr := http.NewRequest("POST", httpUrl, postParam)

	if httpErr != nil {
		LogError("new http post error: ", httpErr.Error())
		return nil, httpErr
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Connection", "Keep-Alive")
	req.Header.Set("Host", host)

	if len(referer) > 0 {
		req.Header.Set("Referer", referer[0])
	}

	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0)")

	if setCookie {
		this.SetHttpCookie(req)
	}

	certFile, _ := filepath.Abs(fmt.Sprintf("certificate/%s/cert.pem", this.CAPath))
	keyFile, _ := filepath.Abs(fmt.Sprintf("certificate/%s/key.unencrypted.pem", this.CAPath))

	cert, certErr := tls.LoadX509KeyPair(certFile, keyFile)

	if certErr != nil {
		LogError("load x509 keypair error: ", certErr.Error())
		return nil, certErr
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
		},
		DisableCompression: true,
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   20 * time.Second,
	}
	res, clientDoErr := client.Do(req)

	if clientDoErr != nil {
		LogError("client do post error: ", clientDoErr.Error())
		return nil, clientDoErr
	}

	if saveCookie {
		this.SaveHttpCookie(res.Cookies(), clearOldCookie)
	}

	return res.Body, nil
}

/**
 * @title 设置http请求cookie
 * @param {*http.Request} req [http请求对象指针]
 */
func (this *SpiderBase) SetHttpCookie(req *http.Request) {
	path, _ := filepath.Abs(fmt.Sprintf("cookies/%s", this.CookiePath))

	cookies := map[string]*http.Cookie{}
	content, readErr := ioutil.ReadFile(path)

	if readErr != nil {
		LogError("cookies read error: ", readErr.Error())
		return
	}

	jsonErr := json.Unmarshal(content, &cookies)

	if jsonErr != nil {
		LogError("cookies json decode error: ", jsonErr.Error())
		return
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
}

/**
 * @title 保存http请求返回的cookie
 * @param {[]*http.Cookie} newCookies [Cookie实例指针]
 * @param {bool} clearOldCookie [是否需要清空原来的cookie]
 */
func (this *SpiderBase) SaveHttpCookie(newCookies []*http.Cookie, clearOldCookie bool) {
	path, _ := filepath.Abs(fmt.Sprintf("cookies/%s", this.CookiePath))

	if len(newCookies) == 0 {
		return
	}

	if clearOldCookie { //清空原cookie，保存新的cookie
		cookies := map[string]*http.Cookie{}

		for _, cookie := range newCookies {
			cookies[cookie.Name] = cookie
		}

		byteArr, jsonErr := json.Marshal(cookies)

		if jsonErr != nil {
			LogError("cookies json encode error: ", jsonErr.Error())
			return
		}

		writeErr := ioutil.WriteFile(path, byteArr, 0777)

		if writeErr != nil {
			LogError("save cookie error: ", writeErr.Error())
		}
	} else { //追加新的cookie
		cookies := map[string]*http.Cookie{}
		content, readErr := ioutil.ReadFile(path)

		if readErr == nil {
			jsonErr := json.Unmarshal(content, &cookies)

			if jsonErr != nil {
				LogError("cookies json decode error: ", jsonErr.Error())
				return
			}
		}

		for _, cookie := range newCookies {
			cookies[cookie.Name] = cookie
		}

		byteArr, jsonErr := json.Marshal(cookies)

		if jsonErr != nil {
			LogError("cookies json encode error: ", jsonErr.Error())
			return
		}

		writeErr := ioutil.WriteFile(path, byteArr, 0777)

		if writeErr != nil {
			LogError("save cookie error: ", writeErr.Error())
		}
	}
}

/**
 * 处理字符串,去除空格字符
 * @param [string] str
 * @return string
 */
func (this *SpiderBase) TrimString(str string) string {
	text := strings.Trim(str, "&nbsp;")
	text = strings.TrimSpace(text)

	return text
}

/**
 * 获取验证码图片
 * @param [string] httpUrl 获取验证码URL
 * @param [string] host 请求头部Host
 * @param {bool} setCookie [请求是否需要加cookie]
 * @param {bool} saveCookie [是否保存返回的cookie]
 * @param {bool} clearOldCookie [是否需要清空原来的cookie]
 * @param [string] imgName 验证码图片保存名称
 * @return [string, error] 返回图片的base64字符串
 */
func (this *SpiderBase) getVerifyCode(httpUrl string, host string, setCookie bool, saveCookie bool, clearOldCookie bool, imgName string) (string, error) {
	resBody, err := this.HttpGet(httpUrl, host, setCookie, saveCookie, clearOldCookie)

	if err != nil {
		LogError("get verifycode error: ", err.Error())
		return "", err
	}

	defer resBody.Close()

	body, readErr := ioutil.ReadAll(resBody)

	if readErr != nil {
		LogError("get verifycode error: ", readErr.Error())
		return "", readErr
	}

	path, _ := filepath.Abs(fmt.Sprintf("verifycode/%s", imgName))
	writeErr := ioutil.WriteFile(path, body, 0777)

	if writeErr != nil {
		LogError("save verifycode error: ", writeErr.Error())
		return "", writeErr
	}

	verifyCodeBase64 := base64.StdEncoding.EncodeToString(body)

	return verifyCodeBase64, nil
}

/**
 * 调用 ShowApi 识别验证码 [showApi是付费服务，需购买可用]
 * @param {string} httpUrl 请求验证码URL
 * @param {string} host 请求的头部Host
 * @param {bool} setCookie [请求是否需要加cookie]
 * @param {bool} saveCookie [是否保存返回的cookie]
 * @param {bool} clearOldCookie [是否需要清空原来的cookie]
 * @param {string} imgName 验证码图片保存名称
 * @param {string} typeId 验证码类型(具体查看showapi文档)
 * @param {string} convertToJpg 是否转化为jpg格式进行识别("0" 否；"1" 是)
 * @return {string} 识别后的验证码字符串
 */
func (this *SpiderBase) CallShowApi(httpUrl string, host string, setCookie bool, saveCookie bool, clearOldCookie bool, imgName string, typeId string, convertToJpg string) (string, error) {
	verifyCodeBase64, err := this.getVerifyCode(httpUrl, host, setCookie, saveCookie, clearOldCookie, imgName)

	if err != nil {
		return "", err
	}

	v := url.Values{}
	v.Set("img_base64", verifyCodeBase64)
	v.Set("typeId", typeId)
	v.Set("convert_to_jpg", convertToJpg)

	postParam := strings.NewReader(v.Encode())
	req, httpErr := http.NewRequest("POST", "http://ali-checkcode.showapi.com/checkcode", postParam)

	if httpErr != nil {
		LogError("call showapi error: ", httpErr.Error())
		return "", httpErr
	}

	appCode := GetConfigString("showapi", "appcode", "794434d1937e4f438223b37fd7951d54")
	req.Header.Set("Authorization", fmt.Sprintf("APPCODE %s", appCode))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	res, clientDoErr := client.Do(req)

	if clientDoErr != nil {
		LogError("call showapi error: ", clientDoErr.Error())
		return "", clientDoErr
	}

	defer res.Body.Close()
	body, readErr := ioutil.ReadAll(res.Body)

	if readErr != nil {
		LogError("call showapi error: ", readErr.Error())
		return "", readErr
	}

	data := &jsons.ShowApiRes{}

	jsonErr := json.Unmarshal(body, &data)

	if jsonErr != nil {
		LogError("call showapi error: ", jsonErr.Error())
		return "", jsonErr
	}

	if data.ShowapiResCode != 0 {
		LogError("call showapi error: ", data.ShowapiResError)
		return "", errors.New(data.ShowapiResError)
	}

	return data.ShowapiResBody.Result, nil
}
