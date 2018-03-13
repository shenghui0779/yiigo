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

// Spider 爬虫基础类 包括：http、https(CA证书)、cookie处理
//
// 做爬虫时需用到另外两个库：
//     1、gbk 转 utf8：gopkg.in/iconv.v1 [https://github.com/qiniu/iconv]
//     2、页面 dom 处理：github.com/PuerkitoBio/goquery
//
// CertPath CA证书存放路径 默认：`certs` 目录，证书需用 `openssl` 转化为 `pem` 格式
// CookiePath cookie存放路径 默认：`cookies` 目录
type Spider struct {
	CertPath   CertPath
	CookiePath string
}

// CertPath cert path
type CertPath struct {
	CertPem           string
	KeyUnencryptedPem string
}

// ReqBody http request body
type ReqBody struct {
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
}

// HTTPGet http get请求
func (s *Spider) HTTPGet(reqBody *ReqBody) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", reqBody.URL, nil)

	if err != nil {
		return nil, err
	}

	s.setHTTPCommonHeader(req, false, reqBody.Host, reqBody.Referer)

	if reqBody.SetCookie {
		err := s.setHTTPCookie(req)

		if err != nil {
			return nil, err
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
		return nil, err
	}

	if reqBody.SaveCookie {
		err := s.saveHTTPCookie(resp.Cookies(), reqBody.CleanOldCookie)

		if err != nil {
			return nil, err
		}
	}

	return resp.Body, nil
}

// HTTPPost http post请求
func (s *Spider) HTTPPost(reqBody *ReqBody) (io.ReadCloser, error) {
	postParam := strings.NewReader(reqBody.PostData.Encode())
	req, err := http.NewRequest("POST", reqBody.URL, postParam)

	if err != nil {
		return nil, err
	}

	s.setHTTPCommonHeader(req, true, reqBody.Host, reqBody.Referer)

	if reqBody.SetCookie {
		err := s.setHTTPCookie(req)

		if err != nil {
			return nil, err
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
		return nil, err
	}

	if reqBody.SaveCookie {
		err := s.saveHTTPCookie(resp.Cookies(), reqBody.CleanOldCookie)

		if err != nil {
			return nil, err
		}
	}

	return resp.Body, nil
}

// HTTPSGet https get请求 CA证书需要用 `openssl` 转换成 `pem` 格式：cert.pem、key.pem
func (s *Spider) HTTPSGet(reqBody *ReqBody) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", reqBody.URL, nil)

	if err != nil {
		return nil, err
	}

	s.setHTTPCommonHeader(req, false, reqBody.Host, reqBody.Referer)

	if reqBody.SetCookie {
		err := s.setHTTPCookie(req)

		if err != nil {
			return nil, err
		}
	}

	certDir := Env.String("spider.certdir", "certs")

	certFile, _ := filepath.Abs(fmt.Sprintf("%s/%s", certDir, s.CertPath.CertPem))
	keyFile, _ := filepath.Abs(fmt.Sprintf("%s/%s", certDir, s.CertPath.KeyUnencryptedPem))

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)

	if err != nil {
		return nil, err
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
		return nil, err
	}

	if reqBody.SaveCookie {
		err := s.saveHTTPCookie(resp.Cookies(), reqBody.CleanOldCookie)

		if err != nil {
			return nil, err
		}
	}

	return resp.Body, nil
}

// HTTPSPost https post请求 CA证书需要用 `openssl` 转换成 `pem` 格式：cert.pem、key.pem
func (s *Spider) HTTPSPost(reqBody *ReqBody) (io.ReadCloser, error) {
	postParam := strings.NewReader(reqBody.PostData.Encode())
	req, err := http.NewRequest("POST", reqBody.URL, postParam)

	if err != nil {
		return nil, err
	}

	s.setHTTPCommonHeader(req, true, reqBody.Host, reqBody.Referer)

	if reqBody.SetCookie {
		err := s.setHTTPCookie(req)

		if err != nil {
			return nil, err
		}
	}

	certDir := Env.String("spider.certdir", "certs")

	certFile, _ := filepath.Abs(fmt.Sprintf("%s/%s", certDir, s.CertPath.CertPem))
	keyFile, _ := filepath.Abs(fmt.Sprintf("%s/%s", certDir, s.CertPath.KeyUnencryptedPem))

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)

	if err != nil {
		return nil, err
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
		return nil, err
	}

	if reqBody.SaveCookie {
		err := s.saveHTTPCookie(resp.Cookies(), reqBody.CleanOldCookie)

		if err != nil {
			return nil, err
		}
	}

	return resp.Body, nil
}

// setHTTPCommonHeader 设置HTTP请求公共头信息
func (s *Spider) setHTTPCommonHeader(req *http.Request, isPost bool, host string, referer string) {
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

// setHTTPCookie 设置http请求cookie
func (s *Spider) setHTTPCookie(req *http.Request) error {
	cookieDir := Env.String("spider.cookiedir", "cookies")
	path, _ := filepath.Abs(fmt.Sprintf("%s/%s", cookieDir, s.CookiePath))

	cookies := map[string]*http.Cookie{}
	content, err := ioutil.ReadFile(path)

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

// saveHTTPCookie 保存http请求返回的cookie
func (s *Spider) saveHTTPCookie(newCookies []*http.Cookie, cleanOldCookie bool) error {
	cookieDir := Env.String("spider.cookiedir", "cookies")
	path, _ := filepath.Abs(fmt.Sprintf("%s/%s", cookieDir, s.CookiePath))

	if len(newCookies) == 0 {
		return nil
	}

	cookies := map[string]*http.Cookie{}

	if cleanOldCookie {
		// 清空原cookie，保存新的cookie
		for _, cookie := range newCookies {
			cookies[cookie.Name] = cookie
		}

		byteArr, err := json.Marshal(cookies)

		if err != nil {
			return err
		}

		err = ioutil.WriteFile(path, byteArr, 0777)

		if err != nil {
			return err
		}
	} else {
		// 追加新的cookie
		content, err := ioutil.ReadFile(path)

		if err != nil {
			return err
		}

		err = json.Unmarshal(content, &cookies)

		if err != nil {
			return err
		}

		for _, cookie := range newCookies {
			cookies[cookie.Name] = cookie
		}

		byteArr, err := json.Marshal(cookies)

		if err != nil {
			return err
		}

		err = ioutil.WriteFile(path, byteArr, 0777)

		if err != nil {
			return err
		}
	}

	return nil
}

// TrimString 处理字符串,去除页面数据中的 `\n` 、`&nbsp;` 和 `空格` 字符
func TrimString(str string) string {
	text := strings.Trim(str, "\n")
	text = strings.Trim(str, "&nbsp;")
	text = strings.TrimSpace(text)

	return text
}
