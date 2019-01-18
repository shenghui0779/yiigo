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

// Spider 爬虫基础类 包括：http、https(SSL证书)、cookie处理
//
// 做爬虫时需用到另外两个库：
//     1、gbk 转 utf8：gopkg.in/iconv.v1 [https://github.com/qiniu/iconv]
//     2、页面 dom 处理：github.com/PuerkitoBio/goquery
//
// HTTPS SSL证书需用 `openssl` 转化为 `pem` 格式：cert.pem、key.pem
type Spider struct {
	client     *http.Client
	cookiePath string
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
		err := s.setSpiderCookie(req)

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
		if err := s.saveSpiderCookie(resp.Cookies(), reqBody.CleanOldCookie); err != nil {
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
		err := s.setSpiderCookie(req)

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
		if err := s.saveSpiderCookie(resp.Cookies(), reqBody.CleanOldCookie); err != nil {
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
func (s *Spider) setSpiderCookie(req *http.Request) error {
	cookies := make(map[string]*http.Cookie)
	content, err := ioutil.ReadFile(s.cookiePath)

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
func (s *Spider) saveSpiderCookie(newCookies []*http.Cookie, cleanOldCookie bool) error {
	if len(newCookies) == 0 {
		return nil
	}

	cookies := make(map[string]*http.Cookie)

	// 追加新的cookie
	if !cleanOldCookie {
		content, err := ioutil.ReadFile(s.cookiePath)

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

	return ioutil.WriteFile(s.cookiePath, b, 0777)
}

// SpiderClientConf 爬虫配置
type SpiderClientConf struct {
	Transport *SpiderTransport
	SSLCert   *SSLCert
}

// buildClient 生成Transport
func (c *SpiderClientConf) buildClient() (*http.Client, error) {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   time.Duration(c.Transport.ConnTimeout) * time.Second,
			KeepAlive: time.Duration(c.Transport.KeepAlive) * time.Second,
			DualStack: true,
		}).DialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // 忽略对服务端传过来的数字证书进行校验
		},
		MaxConnsPerHost:       c.Transport.MaxConnsPerHost,
		MaxIdleConnsPerHost:   c.Transport.MaxIdleConnsPerHost,
		MaxIdleConns:          c.Transport.MaxIdleConns,
		IdleConnTimeout:       time.Duration(c.Transport.IdleConnTimeout) * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// 有SSL证书则处理SSL证书
	if c.SSLCert != nil {
		certFile, err := filepath.Abs(c.SSLCert.CertPem)

		if err != nil {
			return nil, err
		}

		keyFile, err := filepath.Abs(c.SSLCert.KeyUnencryptedPem)

		if err != nil {
			return nil, err
		}

		cert, err := tls.LoadX509KeyPair(certFile, keyFile)

		if err != nil {
			return nil, err
		}

		tr.TLSClientConfig.Certificates = []tls.Certificate{cert}
	}

	return &http.Client{Transport: tr}, nil
}

// SpiderTransport 爬虫 HTTP Transport 配置
type SpiderTransport struct {
	// ConnTimeout 拨号连接超时时间「单位：秒；0：不限」
	ConnTimeout int
	// KeepAlive 连接存活时间「单位：秒；0：短链接」
	KeepAlive int
	// MaxConnsPerHost 每个host最大连接数「0：不限」
	MaxConnsPerHost int
	// MaxIdleConnsPerHost 每个host最大闲置连接数「若为0，则Go默认使用：DefaultMaxIdleConnsPerHost=2」
	MaxIdleConnsPerHost int
	// MaxIdleConns 所有host最大闲置连接数「0：不限」
	MaxIdleConns int
	// IdleConnTimeout 闲置连接的超时时间「单位：秒；0：不限」
	IdleConnTimeout int
}

// SSLCert https ssl cert
type SSLCert struct {
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

// defaultSpiderTransport 爬虫默认Transport
var defaultSpiderTransport = &SpiderTransport{
	ConnTimeout:         30,
	KeepAlive:           60,
	MaxIdleConnsPerHost: 10,
	MaxIdleConns:        100,
	IdleConnTimeout:     60,
}

// NewSpider return a new spider
func NewSpider(cookiePath string, config ...*SpiderClientConf) (*Spider, error) {
	absPath, err := filepath.Abs(cookiePath)

	if err != nil {
		return nil, err
	}

	if err := mkCookieFile(absPath); err != nil {
		return nil, err
	}

	clientConf := &SpiderClientConf{Transport: defaultSpiderTransport}

	if len(config) > 0 {
		conf := config[0]

		if conf.Transport != nil {
			clientConf.Transport = conf.Transport
		}

		if conf.SSLCert != nil {
			clientConf.SSLCert = conf.SSLCert
		}
	}

	c, err := clientConf.buildClient()

	if err != nil {
		return nil, err
	}

	spider := &Spider{
		client:     c,
		cookiePath: absPath,
	}

	return spider, nil
}

// mkCookieFile 创建cookie文件
func mkCookieFile(path string) error {
	dir := filepath.Dir(path)

	// 目录不存在则创建目录
	if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	// 文件不存在则创建文件
	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		if _, err := os.Create(path); err != nil {
			return err
		}
	}

	return nil
}
