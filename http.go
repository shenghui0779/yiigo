package yiigo

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// YiiClient HTTP request client
var YiiClient = &http.Client{
	Timeout: 5 * time.Second,
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

	// custom timeout
	if len(timeout) > 0 {
		YiiClient.Timeout = timeout[0]
	}

	resp, err := YiiClient.Do(req)

	if err != nil {
		return nil, err
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

// HTTPPost http post request
func HTTPPost(url string, body []byte, headers map[string]string, timeout ...time.Duration) ([]byte, error) {
	reader := bytes.NewReader(body)

	req, err := http.NewRequest("POST", url, reader)

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

	// custom timeout
	if len(timeout) > 0 {
		YiiClient.Timeout = timeout[0]
	}

	resp, err := YiiClient.Do(req)

	if err != nil {
		return nil, err
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
