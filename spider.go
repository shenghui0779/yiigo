package yiigo

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
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
	// Headers 请求头
	Headers map[string]string
	// PostData POST请求数据
	PostData []byte
	// NeedSetCookie 请求是否需要带上cookie
	NeedSetCookie bool
	// NeedSaveCookie 是否需要保存返回的cookie
	NeedSaveCookie bool
	// ClearOldCookie 是否需要清空原来的cookie
	CleanOldCookie bool
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
		MaxConnsPerHost:       20,
		MaxIdleConnsPerHost:   10,
		MaxIdleConns:          100,
		IdleConnTimeout:       60 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
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
	if len(reqBody.Headers) != 0 {
		for k, v := range reqBody.Headers {
			req.Header.Set(k, v)
		}
	}

	// 请求带上cookie
	if reqBody.NeedSetCookie {
		err := setSpiderCookie(req, s.cookiePath)

		if err != nil {
			return nil, err
		}
	}

	// 设置超时时间
	t := httpDefaultTimeout

	if reqBody.Timeout != 0 {
		t = reqBody.Timeout
	}

	ctx, cancel := context.WithTimeout(context.TODO(), t)

	defer cancel()

	resp, err := s.client.Do(req.WithContext(ctx))

	if err != nil {
		return nil, err
	}

	// 保存新的cookie
	if reqBody.NeedSaveCookie {
		if err := saveSpiderCookie(resp.Cookies(), s.cookiePath, reqBody.CleanOldCookie); err != nil {
			return nil, err
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(ioutil.Discard, resp.Body)

		return nil, fmt.Errorf("error http code: %d", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return b, nil
}

// HTTPPost http post请求
func (s *Spider) HTTPPost(reqBody *SpiderReqBody) ([]byte, error) {
	req, err := http.NewRequest("POST", reqBody.URL, bytes.NewReader(reqBody.PostData))

	if err != nil {
		return nil, err
	}

	// 设置请求头
	if len(reqBody.Headers) != 0 {
		for k, v := range reqBody.Headers {
			req.Header.Set(k, v)
		}
	}

	// 请求带上cookie
	if reqBody.NeedSetCookie {
		err := setSpiderCookie(req, s.cookiePath)

		if err != nil {
			return nil, err
		}
	}

	// 设置超时时间
	t := httpDefaultTimeout

	if reqBody.Timeout != 0 {
		t = reqBody.Timeout
	}

	ctx, cancel := context.WithTimeout(context.TODO(), t)

	defer cancel()

	resp, err := s.client.Do(req.WithContext(ctx))

	if err != nil {
		return nil, err
	}

	// 保存新的cookie
	if reqBody.NeedSaveCookie {
		if err := saveSpiderCookie(resp.Cookies(), s.cookiePath, reqBody.CleanOldCookie); err != nil {
			return nil, err
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(ioutil.Discard, resp.Body)

		return nil, fmt.Errorf("error http code: %d", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return b, nil
}

// setSpiderCookie 设置http请求cookie
func setSpiderCookie(req *http.Request, cookiePath string) error {
	cookies := make(map[string]*http.Cookie)
	content, err := ioutil.ReadFile(cookiePath)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(content, &cookies); err != nil {
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

	cookies := make(map[string]*http.Cookie)

	// 追加新的cookie
	if !cleanOldCookie {
		content, err := ioutil.ReadFile(cookiePath)

		if err != nil {
			return err
		}

		if len(content) > 0 {
			if err := json.Unmarshal(content, &cookies); err != nil {
				return err
			}
		}
	}

	for _, cookie := range newCookies {
		cookies[cookie.Name] = cookie
	}

	b, err := json.Marshal(cookies)

	if err != nil {
		return err
	}

	return ioutil.WriteFile(cookiePath, b, 0777)
}
