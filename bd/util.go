package bd

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type Request struct {
	Host     string
	Route    string
	QueryArg interface{}
	Body     interface{}
	Headers  map[string]string
}

func DoHTTPRequest(url string, body io.Reader, headers map[string]string) (string, int, error) {
	timeout := 5 * time.Second
	retryTimes := 3
	tr := &http.Transport{
		MaxIdleConnsPerHost: -1,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}
	httpClient.Timeout = timeout
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", 0, err
	}
	// request header
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	var resp *http.Response
	for i := 1; i <= retryTimes; i++ {
		resp, err = httpClient.Do(req)
		if err == nil {
			break
		}
		if i == retryTimes {
			return "", 0, err
		}
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", resp.StatusCode, err
	}
	return string(respBody), resp.StatusCode, nil
}

// for superfile2
func SendHTTPRequest(url string, body io.Reader, headers map[string]string) (string, int, error) {
	timeout := 60 * time.Second
	retryTimes := 3
	postData, _ := ioutil.ReadAll(body)
	var resp *http.Response
	for i := 1; i <= retryTimes; i++ {
		tr := &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			MaxIdleConnsPerHost: -1,
		}
		httpClient := &http.Client{Transport: tr}
		httpClient.Timeout = timeout
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(postData))
		if err != nil {
			return "", 0, err
		}
		// request header
		for k, v := range headers {
			req.Header.Add(k, v)
		}
		resp, err = httpClient.Do(req)
		if err == nil {
			break
		}
		if i == retryTimes {
			return "", 0, err
		}
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", resp.StatusCode, err
	}

	return string(respBody), resp.StatusCode, nil
}

// for download
func Do2HTTPRequest(ctx context.Context, url string, body io.Reader, headers map[string]string) (io.ReadCloser, int, error) {
	// timeout := 500 * time.Second
	retryTimes := 3
	tr := &http.Transport{
		MaxIdleConnsPerHost: -1,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}
	// httpClient.Timeout = timeout
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, 0, err
	}
	// request header
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	var resp *http.Response
	for i := 1; i <= retryTimes; i++ {
		resp, err = httpClient.Do(req)
		if err == nil {
			break
		}
		if i == retryTimes {
			return nil, 0, err
		}
	}

	return resp.Body, resp.StatusCode, nil
}
