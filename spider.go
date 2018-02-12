package yiigo

import (
	"crypto/tls"
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

// 爬虫基础类 [包括：http、https(CA证书)、cookie、验证码处理]
// 做爬虫时需用到另外两个库：
//     1、gbk 转 utf8：gopkg.in/iconv.v1 [https://github.com/qiniu/iconv]
//     2、页面 dom 处理：github.com/PuerkitoBio/goquery
// CertPath {CertPath} CA证书存放路径 [默认 certs 目录，证书需用 openssl 转化为 pem格式]
// CookiePath {string} cookie存放路径 [默认 cookies 目录]

// Spider spider
type Spider struct {
	CertPath   CertPath
	CookiePath string
}

// CertPath cert path
type CertPath struct {
	CertPem           string
	KeyUnencryptedPem string
}

// HTTPGet http get请求
// @param httpURL string 请求地址
// @param host string 请求头部 Host
// @param setCookie bool 请求是否需要加 cookie
// @param saveCookie bool 是否保存返回的 cookie
// @param clearOldCookie bool 是否需要清空原来的 cookie
// @param referer string 请求头部 referer
// @return io.ReadCloser
func (s *Spider) HTTPGet(httpURL string, host string, setCookie bool, saveCookie bool, clearOldCookie bool, referer ...string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", httpURL, nil)

	if err != nil {
		return nil, fmt.Errorf("[Spider] %v", err)
	}

	s.setHTTPCommonHeader(req, false, host, referer...)

	if setCookie {
		err := s.setHTTPCookie(req)

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
		err := s.saveHTTPCookie(resp.Cookies(), clearOldCookie)

		if err != nil {
			return nil, fmt.Errorf("[Spider] %v", err)
		}
	}

	return resp.Body, nil
}

// HTTPPost http post请求
// @param httpURL string 请求地址
// @param host string 请求头部 Host
// @param v url.Values post参数
// @param setCookie bool 请求是否需要加 cookie
// @param saveCookie bool 是否保存返回的 cookie
// @param clearOldCookie bool 是否需要清空原来的 cookie
// @param referer string 请求头部 referer
// @return io.ReadCloser
func (s *Spider) HTTPPost(httpURL string, host string, v url.Values, setCookie bool, saveCookie bool, clearOldCookie bool, referer ...string) (io.ReadCloser, error) {
	postParam := strings.NewReader(v.Encode())
	req, err := http.NewRequest("POST", httpURL, postParam)

	if err != nil {
		return nil, fmt.Errorf("[Spider] %v", err)
	}

	s.setHTTPCommonHeader(req, true, host, referer...)

	if setCookie {
		err := s.setHTTPCookie(req)

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
		err := s.saveHTTPCookie(resp.Cookies(), clearOldCookie)

		if err != nil {
			return nil, fmt.Errorf("[Spider] %v", err)
		}
	}

	return resp.Body, nil
}

// HTTPSGet https get请求 [https 需要CA证书，用 openssl 转换成 pem格式：cert.pem、key.pem]
// @param httpURL string 请求地址
// @param host string 请求头部 Host
// @param setCookie bool 请求是否需要加 cookie
// @param saveCookie bool 是否保存返回的 cookie
// @param clearOldCookie bool 是否需要清空原来的 cookie
// @param referer string 请求头部 referer
// @return io.ReadCloser
func (s *Spider) HTTPSGet(httpURL string, host string, setCookie bool, saveCookie bool, clearOldCookie bool, referer ...string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", httpURL, nil)

	if err != nil {
		return nil, fmt.Errorf("[Spider] %v", err)
	}

	s.setHTTPCommonHeader(req, false, host, referer...)

	if setCookie {
		err := s.setHTTPCookie(req)

		if err != nil {
			return nil, fmt.Errorf("[Spider] %v", err)
		}
	}

	certDir := Env.String("spider.certdir", "certs")

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
		err := s.saveHTTPCookie(resp.Cookies(), clearOldCookie)

		if err != nil {
			return nil, fmt.Errorf("[Spider] %v", err)
		}
	}

	return resp.Body, nil
}

// HTTPSPost https post请求 [https 需要CA证书，用openssl转换成pem格式：cert.pem、key.pem]
// @param httpURL string 请求地址
// @param host string 请求头部Host
// @param v url.Values post参数
// @param setCookie bool 请求是否需要加cookie
// @param saveCookie bool 是否保存返回的cookie
// @param clearOldCookie bool 是否需要清空原来的cookie
// @param referer string 请求头部referer
// @return io.ReadCloser
func (s *Spider) HTTPSPost(httpURL string, host string, v url.Values, setCookie bool, saveCookie bool, clearOldCookie bool, referer ...string) (io.ReadCloser, error) {
	postParam := strings.NewReader(v.Encode())
	req, err := http.NewRequest("POST", httpURL, postParam)

	if err != nil {
		return nil, fmt.Errorf("[Spider] %v", err)
	}

	s.setHTTPCommonHeader(req, true, host, referer...)

	if setCookie {
		err := s.setHTTPCookie(req)

		if err != nil {
			return nil, fmt.Errorf("[Spider] %v", err)
		}
	}

	certDir := Env.String("spider.certdir", "certs")

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
		err := s.saveHTTPCookie(resp.Cookies(), clearOldCookie)

		if err != nil {
			return nil, fmt.Errorf("[Spider] %v", err)
		}
	}

	return resp.Body, nil
}

// setHTTPCommonHeader 设置HTTP请求公共头部
// @param req http.Request http请求对象指针
// @param isPost bool 是否为post请求
// @param host string 请求头部Host
// @param referer string 请求头部referer
func (s *Spider) setHTTPCommonHeader(req *http.Request, isPost bool, host string, referer ...string) {
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/,;q=0.8")
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

// setHTTPCookie 设置http请求cookie
// @param req http.Request http请求对象指针
// @return error
func (s *Spider) setHTTPCookie(req *http.Request) error {
	cookieDir := Env.String("spider.cookiedir", "cookies")
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

// saveHTTPCookie 保存http请求返回的cookie
// @param newCookies []http.Cookie Cookie实例指针
// @param clearOldCookie bool 是否需要清空原来的cookie
// @param error
func (s *Spider) saveHTTPCookie(newCookies []*http.Cookie, clearOldCookie bool) error {
	cookieDir := Env.String("spider.cookiedir", "cookies")
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

// TrimString 处理字符串,去除页面数据中的 "\n" 、"&nbsp;" 和 空格字符
// @param str string
// @return string
func TrimString(str string) string {
	text := strings.Trim(str, "\n")
	text = strings.Trim(str, "&nbsp;")
	text = strings.TrimSpace(text)

	return text
}
