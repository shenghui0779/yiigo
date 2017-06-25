package yiigo

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

/**
 * 爬虫基础类 [包括：http、https(CA证书)、cookie、验证码处理]
 * 做爬虫时需用到另外两个库：
 * 1、gbk 转 utf8：gopkg.in/iconv.v1 [https://github.com/qiniu/iconv]
 * 2、页面 dom 处理：github.com/PuerkitoBio/goquery
 * CertPath {*CertPath} CA证书存放路径 [默认 certs 目录，证书需用 openssl 转化为 pem格式]
 * CookiePath {string} cookie存放路径 [默认 cookies 目录]
 * 验证码图片默认存放路径为 captcha 目录
 */

type SpiderBase struct {
	CertPath   *CertPath
	CookiePath string
}

type CertPath struct {
	CertPem           string
	KeyUnencryptedPem string
}

// 验证码接口返回
type showAPIRes struct {
	ShowapiResCode  int          `json:"showapi_res_code"`
	ShowapiResError string       `json:"showapi_res_error"`
	ShowapiResBody  *showAPIBody `json:"showapi_res_body"`
}

type showAPIBody struct {
	Result  string `json:"Result"`
	RetCode int    `json:"ret_code"`
	ID      string `json:"Id"`
}

/**
 * HttpGet请求
 * @param httpUrl string 请求地址
 * @param host string 请求头部 Host
 * @param setCookie bool 请求是否需要加 cookie
 * @param saveCookie bool 是否保存返回的 cookie
 * @param clearOldCookie bool 是否需要清空原来的 cookie
 * @param referer string 请求头部 referer
 * @return io.ReadCloser
 */
func (s *SpiderBase) HttpGet(httpUrl string, host string, setCookie bool, saveCookie bool, clearOldCookie bool, referer ...string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", httpUrl, nil)

	if err != nil {
		return nil, fmt.Errorf("[Spider] %v", err)
	}

	s.setHttpCommonHeader(req, false, host, referer...)

	if setCookie {
		err := s.setHttpCookie(req)

		if err != nil {
			return nil, fmt.Errorf("[Spider] %v", err)
		}
	}

	//忽略对服务端传过来的数字证书进行校验
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   20 * time.Second,
	}
	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("[Spider] %v", err)
	}

	if saveCookie {
		err := s.saveHttpCookie(resp.Cookies(), clearOldCookie)

		if err != nil {
			return nil, fmt.Errorf("[Spider] %v", err)
		}
	}

	return resp.Body, nil
}

/**
 * HttpPost请求
 * @param httpUrl string 请求地址
 * @param host string 请求头部 Host
 * @param v url.Values post参数
 * @param setCookie bool 请求是否需要加 cookie
 * @param saveCookie bool 是否保存返回的 cookie
 * @param clearOldCookie bool 是否需要清空原来的 cookie
 * @param referer string 请求头部 referer
 * @return io.ReadCloser
 */
func (s *SpiderBase) HttpPost(httpUrl string, host string, v url.Values, setCookie bool, saveCookie bool, clearOldCookie bool, referer ...string) (io.ReadCloser, error) {
	postParam := strings.NewReader(v.Encode())
	req, err := http.NewRequest("POST", httpUrl, postParam)

	if err != nil {
		return nil, fmt.Errorf("[Spider] %v", err)
	}

	s.setHttpCommonHeader(req, true, host, referer...)

	if setCookie {
		err := s.setHttpCookie(req)

		if err != nil {
			return nil, fmt.Errorf("[Spider] %v", err)
		}
	}

	//忽略对服务端传过来的数字证书进行校验
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   20 * time.Second,
	}
	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("[Spider] %v", err)
	}

	if saveCookie {
		err := s.saveHttpCookie(resp.Cookies(), clearOldCookie)

		if err != nil {
			return nil, fmt.Errorf("[Spider] %v", err)
		}
	}

	return resp.Body, nil
}

/**
 * HttpsGet请求 [https 需要CA证书，用 openssl 转换成 pem格式：cert.pem、key.pem]
 * @param httpUrl string 请求地址
 * @param host string 请求头部 Host
 * @param setCookie bool 请求是否需要加 cookie
 * @param saveCookie bool 是否保存返回的 cookie
 * @param clearOldCookie bool 是否需要清空原来的 cookie
 * @param referer string 请求头部 referer
 * @return io.ReadCloser
 */
func (s *SpiderBase) HttpsGet(httpUrl string, host string, setCookie bool, saveCookie bool, clearOldCookie bool, referer ...string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", httpUrl, nil)

	if err != nil {
		return nil, fmt.Errorf("[Spider] %v", err)
	}

	s.setHttpCommonHeader(req, false, host, referer...)

	if setCookie {
		err := s.setHttpCookie(req)

		if err != nil {
			return nil, fmt.Errorf("[Spider] %v", err)
		}
	}

	certDir := GetEnvString("spider", "certdir", "certs")

	certFile, _ := filepath.Abs(fmt.Sprintf("%s/%s", certDir, s.CertPath.CertPem))
	keyFile, _ := filepath.Abs(fmt.Sprintf("%s/%s", certDir, s.CertPath.KeyUnencryptedPem))

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)

	if err != nil {
		return nil, fmt.Errorf("[Spider] %v", err)
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
	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("[Spider] %v", err)
	}

	if saveCookie {
		err := s.saveHttpCookie(resp.Cookies(), clearOldCookie)

		if err != nil {
			return nil, fmt.Errorf("[Spider] %v", err)
		}
	}

	return resp.Body, nil
}

/**
 * post请求 [https 需要CA证书，用openssl转换成pem格式：cert.pem、key.pem]
 * @param httpUrl string 请求地址
 * @param host string 请求头部Host
 * @param v url.Values post参数
 * @param setCookie bool 请求是否需要加cookie
 * @param saveCookie bool 是否保存返回的cookie
 * @param clearOldCookie bool 是否需要清空原来的cookie
 * @param referer string 请求头部referer
 * @return io.ReadCloser
 */
func (s *SpiderBase) HttpsPost(httpUrl string, host string, v url.Values, setCookie bool, saveCookie bool, clearOldCookie bool, referer ...string) (io.ReadCloser, error) {
	postParam := strings.NewReader(v.Encode())
	req, err := http.NewRequest("POST", httpUrl, postParam)

	if err != nil {
		return nil, fmt.Errorf("[Spider] %v", err)
	}

	s.setHttpCommonHeader(req, true, host, referer...)

	if setCookie {
		err := s.setHttpCookie(req)

		if err != nil {
			return nil, fmt.Errorf("[Spider] %v", err)
		}
	}

	certDir := GetEnvString("spider", "certdir", "certs")

	certFile, _ := filepath.Abs(fmt.Sprintf("%s/%s", certDir, s.CertPath.CertPem))
	keyFile, _ := filepath.Abs(fmt.Sprintf("%s/%s", certDir, s.CertPath.KeyUnencryptedPem))

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)

	if err != nil {
		return nil, fmt.Errorf("[Spider] %v", err)
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
	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("[Spider] %v", err)
	}

	if saveCookie {
		err := s.saveHttpCookie(resp.Cookies(), clearOldCookie)

		if err != nil {
			return nil, fmt.Errorf("[Spider] %v", err)
		}
	}

	return resp.Body, nil
}

/**
 * 设置Http请求公共头部
 * @param req *http.Request http请求对象指针
 * @param isPost bool 是否为post请求
 * @param host string 请求头部Host
 * @param referer string 请求头部referer
 */
func (s *SpiderBase) setHttpCommonHeader(req *http.Request, isPost bool, host string, referer ...string) {
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8")

	if isPost {
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req.Header.Set("Cache-Control", "max-age=0")
	}

	req.Header.Set("Connection", "Keep-Alive")
	req.Header.Set("Host", host)

	if len(referer) > 0 {
		req.Header.Set("Referer", referer[0])
	}

	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0)")
}

/**
 * 设置http请求cookie
 * @param req *http.Request http请求对象指针
 * @return error
 */
func (s *SpiderBase) setHttpCookie(req *http.Request) error {
	cookieDir := GetEnvString("spider", "cookiedir", "cookies")
	path, _ := filepath.Abs(fmt.Sprintf("%s/%s", cookieDir, s.CookiePath))

	cookies := map[string]*http.Cookie{}
	content, err := ioutil.ReadFile(path)

	if err != nil {
		return fmt.Errorf("[Spider] %v", err)
	}

	err = json.Unmarshal(content, &cookies)

	if err != nil {
		return fmt.Errorf("[Spider] %v", err)
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	return nil
}

/**
 * 保存http请求返回的cookie
 * @param newCookies []*http.Cookie Cookie实例指针
 * @param clearOldCookie bool 是否需要清空原来的cookie
 * @param error
 */
func (s *SpiderBase) saveHttpCookie(newCookies []*http.Cookie, clearOldCookie bool) error {
	cookieDir := GetEnvString("spider", "cookiedir", "cookies")
	path, _ := filepath.Abs(fmt.Sprintf("%s/%s", cookieDir, s.CookiePath))

	if len(newCookies) == 0 {
		return nil
	}

	if clearOldCookie { //清空原cookie，保存新的cookie
		cookies := map[string]*http.Cookie{}

		for _, cookie := range newCookies {
			cookies[cookie.Name] = cookie
		}

		byteArr, err := json.Marshal(cookies)

		if err != nil {
			return fmt.Errorf("[Spider] %v", err)
		}

		err = ioutil.WriteFile(path, byteArr, 0777)

		if err != nil {
			return fmt.Errorf("[Spider] %v", err)
		}
	} else { //追加新的cookie
		cookies := map[string]*http.Cookie{}
		content, err := ioutil.ReadFile(path)

		if err != nil {
			return fmt.Errorf("[Spider] %v", err)
		}

		err = json.Unmarshal(content, &cookies)

		if err != nil {
			return fmt.Errorf("[Spider] %v", err)
		}

		for _, cookie := range newCookies {
			cookies[cookie.Name] = cookie
		}

		byteArr, err := json.Marshal(cookies)

		if err != nil {
			return fmt.Errorf("[Spider] %v", err)
		}

		err = ioutil.WriteFile(path, byteArr, 0777)

		if err != nil {
			return fmt.Errorf("[Spider] %v", err)
		}
	}

	return nil
}

/**
 * 获取验证码图片 (base64字符串)
 * @param httpUrl string 获取验证码URL
 * @param host string 请求头部Host
 * @param setCookie bool 请求是否需要加cookie
 * @param saveCookie bool 是否保存返回的cookie
 * @param clearOldCookie bool 是否需要清空原来的cookie
 * @param captchaImg string 验证码图片保存名称
 * @return string, error
 */
func (s *SpiderBase) getCaptchaBase64(httpUrl string, host string, setCookie bool, saveCookie bool, clearOldCookie bool, captchaImg string) (string, error) {
	resBody, err := s.HttpGet(httpUrl, host, setCookie, saveCookie, clearOldCookie)

	if err != nil {
		return "", fmt.Errorf("[Spider] %v", err)
	}

	defer resBody.Close()

	body, err := ioutil.ReadAll(resBody)

	if err != nil {
		return "", fmt.Errorf("[Spider] %v", err)
	}

	verifyDir := GetEnvString("spider", "captchadir", "captcha")

	path, _ := filepath.Abs(fmt.Sprintf("%s/%s", verifyDir, captchaImg))
	err = ioutil.WriteFile(path, body, 0777)

	if err != nil {
		return "", fmt.Errorf("[Spider] %v", err)
	}

	captchaBase64 := base64.StdEncoding.EncodeToString(body)

	return captchaBase64, nil
}

/**
 * 调用 showapi 接口识别验证码 [showApi是付费服务：https://market.aliyun.com/products/57124001/cmapi011148.html#sku=yuncode514800004]
 * @param httpUrl string 请求验证码URL
 * @param host string 请求的头部Host
 * @param setCookie bool 请求是否需要加cookie
 * @param saveCookie bool 是否保存返回的cookie
 * @param clearOldCookie bool 是否需要清空原来的cookie
 * @param captchaImg string 验证码图片保存名称
 * @param typeId string 验证码类型 (具体查看showapi文档)
 * @param convertToJpg string 是否转化为jpg格式进行识别("0" 否；"1" 是)
 * @return string
 */
func (s *SpiderBase) GetCaptchaCode(httpUrl string, host string, setCookie bool, saveCookie bool, clearOldCookie bool, captchaImg string, typeId string, convertToJpg string) (string, error) {
	captchaBase64, err := s.getCaptchaBase64(httpUrl, host, setCookie, saveCookie, clearOldCookie, captchaImg)

	if err != nil {
		return "", err
	}

	v := url.Values{}

	v.Set("img_base64", captchaBase64)
	v.Set("typeId", typeId)
	v.Set("convert_to_jpg", convertToJpg)

	postParam := strings.NewReader(v.Encode())
	req, err := http.NewRequest("POST", "http://ali-checkcode.showapi.com/checkcode", postParam)

	if err != nil {
		return "", fmt.Errorf("[Spider] %v", err)
	}

	appCode := GetEnvString("spider", "appcode", "794434d1937e4f438223b37fd7951d54")
	req.Header.Set("Authorization", fmt.Sprintf("APPCODE %s", appCode))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return "", fmt.Errorf("[Spider] %v", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", fmt.Errorf("[Spider] %v", err)
	}

	data := &showAPIRes{}

	err = json.Unmarshal(body, &data)

	if err != nil {
		return "", fmt.Errorf("[Spider] %v", err)
	}

	if data.ShowapiResCode != 0 {
		return "", fmt.Errorf("[Spider] %v", data.ShowapiResError)
	}

	return data.ShowapiResBody.Result, nil
}

/**
 * 处理字符串,去除页面数据中的 "\n" 、"&nbsp;" 和 空格字符
 * @param str string
 * @return string
 */
func TrimString(str string) string {
	text := strings.Trim(str, "\n")
	text = strings.Trim(str, "&nbsp;")
	text = strings.TrimSpace(text)

	return text
}
