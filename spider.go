package yiigo

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Spider 爬虫基础类 包括：http、https(CA证书)、cookie处理
//
// 做爬虫时需用到另外两个库：
//     1、gbk 转 utf8：gopkg.in/iconv.v1 [https://github.com/qiniu/iconv]
//     2、页面 dom 处理：github.com/PuerkitoBio/goquery
//
// HTTPS CA证书需用 `openssl` 转化为 `pem` 格式：cert.pem、key.pem
type Spider struct {
	client     *http.Client
	cookiePath string
}

// HTTPSCert https cert
type HTTPSCert struct {
	CertPem           string
	KeyUnencryptedPem string
}

// SpiderReqBody spider http request body
type SpiderReqBody struct {
	// URL 请求地址
	URL string
	// Host 请求头Host
	Host string
	// PostData post参数
	PostData url.Values
	// SetCookie 请求是否需要加cookie
	SetCookie bool
	// SaveCookie 是否保存返回的cookie
	SaveCookie bool
	// ClearOldCookie 是否需要清空原来的cookie
	CleanOldCookie bool
	// Referer 请求头Referer
	Referer string
	// Timeout 超时时间
	Timeout time.Duration
}

// NewSpider return a new spider
func NewSpider(cookieFile string, cert *HTTPSCert) (*Spider, error) {
	cookiePath, err := filepath.Abs(cookieFile)

	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(cookiePath)

	// 目录不存在则创建目录
	if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return nil, err
		}
	}

	// 文件不存在则创建文件
	if _, err := os.Stat(cookiePath); err != nil && os.IsNotExist(err) {
		if _, err := os.Create(cookiePath); err != nil {
			return nil, err
		}
	}

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 60 * time.Second,
			DualStack: true,
		}).DialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 忽略对服务端传过来的数字证书进行校验
		},
		MaxConnsPerHost:       200,
		MaxIdleConnsPerHost:   100,
		MaxIdleConns:          100,
		IdleConnTimeout:       60 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableCompression:    true,
	}

	// 有CA证书则处理CA证书
	if cert != nil {
		certFile, err := filepath.Abs(cert.CertPem)

		if err != nil {
			return nil, err
		}

		keyFile, err := filepath.Abs(cert.KeyUnencryptedPem)

		if err != nil {
			return nil, err
		}

		cert, err := tls.LoadX509KeyPair(certFile, keyFile)

		if err != nil {
			return nil, err
		}

		tr.TLSClientConfig.Certificates = []tls.Certificate{cert}
	}

	spider := &Spider{
		client: &http.Client{
			Transport: tr,
			Timeout:   10 * time.Second,
		},
		cookiePath: cookiePath,
	}

	return spider, nil
}

// HTTPGet http get请求
func (s *Spider) HTTPGet(reqBody *SpiderReqBody) ([]byte, error) {
	req, err := http.NewRequest("GET", reqBody.URL, nil)

	if err != nil {
		return nil, err
	}

	// 设置请求头
	setSpiderHeader(req, false, reqBody.Host, reqBody.Referer)

	// 请求带上cookie
	if reqBody.SetCookie {
		err := setSpiderCookie(req, s.cookiePath)

		if err != nil {
			return nil, err
		}
	}

	// 设置超时时间
	if reqBody.Timeout != 0 {
		s.client.Timeout = reqBody.Timeout
	}

	resp, err := s.client.Do(req)

	if err != nil {
		return nil, err
	}

	// 保存新的cookie
	if reqBody.SaveCookie {
		err := saveSpiderCookie(resp.Cookies(), s.cookiePath, reqBody.CleanOldCookie)

		if err != nil {
			return nil, err
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error http code: %d", resp.StatusCode)
	}

	var b []byte

	if resp.Body == http.NoBody {
		return b, nil
	}

	b, err = ioutil.ReadAll(resp.Body)

	return b, err
}

// HTTPPost http post请求
func (s *Spider) HTTPPost(reqBody *SpiderReqBody) ([]byte, error) {
	postParam := strings.NewReader(reqBody.PostData.Encode())
	req, err := http.NewRequest("POST", reqBody.URL, postParam)

	if err != nil {
		return nil, err
	}

	// 设置请求头
	setSpiderHeader(req, true, reqBody.Host, reqBody.Referer)

	// 请求带上cookie
	if reqBody.SetCookie {
		err := setSpiderCookie(req, s.cookiePath)

		if err != nil {
			return nil, err
		}
	}

	// 设置超时时间
	if reqBody.Timeout != 0 {
		s.client.Timeout = reqBody.Timeout
	}

	resp, err := s.client.Do(req)

	if err != nil {
		return nil, err
	}

	// 保存新的cookie
	if reqBody.SaveCookie {
		err := saveSpiderCookie(resp.Cookies(), s.cookiePath, reqBody.CleanOldCookie)

		if err != nil {
			return nil, err
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error http code: %d", resp.StatusCode)
	}

	var b []byte

	if resp.Body == http.NoBody {
		return b, nil
	}

	b, err = ioutil.ReadAll(resp.Body)

	return b, err
}

// setSpiderHeader 设置HTTP请求公共头信息
func setSpiderHeader(req *http.Request, isPost bool, host string, referer string) {
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

	if referer != "" {
		req.Header.Set("Referer", referer)
	}

	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0)")
}

// setSpiderCookie 设置http请求cookie
func setSpiderCookie(req *http.Request, cookiePath string) error {
	cookies := map[string]*http.Cookie{}
	content, err := ioutil.ReadFile(cookiePath)

	if err != nil {
		return err
	}

	err = json.Unmarshal(content, &cookies)

	if err != nil {
		return err
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	return nil
}

// saveSpiderCookie 保存http请求返回的cookie
func saveSpiderCookie(newCookies []*http.Cookie, cookiePath string, cleanOldCookie bool) error {
	if len(newCookies) == 0 {
		return nil
	}

	cookies := map[string]*http.Cookie{}

	// 追加新的cookie
	if !cleanOldCookie {
		content, err := ioutil.ReadFile(cookiePath)

		if err != nil {
			return err
		}

		if len(content) > 0 {
			err = json.Unmarshal(content, &cookies)

			if err != nil {
				return err
			}
		}
	}

	for _, cookie := range newCookies {
		cookies[cookie.Name] = cookie
	}

	byteArr, err := json.Marshal(cookies)

	if err != nil {
		return err
	}

	return ioutil.WriteFile(cookiePath, byteArr, 0777)
}
