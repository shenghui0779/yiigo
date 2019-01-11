package yiigo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type httpConf struct {
	ConnTimeout         int `toml:"connTimeout"`
	KeepAlive           int `toml:"keepAlive"`
	MaxConnsPerHost     int `toml:"maxConnsPerHost"`
	MaxIdleConnsPerHost int `toml:"maxIdleConnsPerHost"`
	MaxIdleConns        int `toml:"maxIdleConns"`
	IdleConnTimeout     int `toml:"idleConnTimeout"`
}

// httpDefaultTimeout HTTP request default timeout
const httpDefaultTimeout = 10 * time.Second

// httpClient HTTP request client
var httpClient *http.Client

func initHTTPClient() {
	conf := &httpConf{
		ConnTimeout:         30,
		KeepAlive:           60,
		MaxIdleConnsPerHost: 10,
		MaxIdleConns:        100,
		IdleConnTimeout:     60,
	}

	Env.Unmarshal("http", conf)

	httpClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   time.Duration(conf.ConnTimeout) * time.Second,
				KeepAlive: time.Duration(conf.KeepAlive) * time.Second,
				DualStack: true,
			}).DialContext,
			MaxConnsPerHost:       conf.MaxConnsPerHost,
			MaxIdleConnsPerHost:   conf.MaxIdleConnsPerHost,
			MaxIdleConns:          conf.MaxIdleConns,
			IdleConnTimeout:       time.Duration(conf.IdleConnTimeout) * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

// HTTPGet http get request
func HTTPGet(url string, headers map[string]string, timeout ...time.Duration) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	// custom headers
	if len(headers) != 0 {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	t := httpDefaultTimeout

	// custom timeout
	if len(timeout) > 0 {
		t = timeout[0]
	}

	ctx, cancel := context.WithTimeout(context.TODO(), t)

	defer cancel()

	resp, err := httpClient.Do(req.WithContext(ctx))

	if err != nil {
		return nil, err
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

// HTTPPost http post request, the default `content-type` is 'application/json'.
func HTTPPost(url string, body []byte, headers map[string]string, timeout ...time.Duration) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	// custom headers
	if len(headers) != 0 {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	t := httpDefaultTimeout

	// custom timeout
	if len(timeout) > 0 {
		t = timeout[0]
	}

	ctx, cancel := context.WithTimeout(context.TODO(), t)

	defer cancel()

	resp, err := httpClient.Do(req.WithContext(ctx))

	if err != nil {
		return nil, err
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
