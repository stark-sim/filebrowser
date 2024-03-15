package http

import (
	"encoding/json"
	"fmt"
	"github.com/r3labs/sse/v2"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

var DirInfoURL = os.Getenv("USER_CENTER_HOST") + "/v1/cloud-files"

type downloadInput struct {
	MD5      string `json:"md5"`
	Target   string `json:"target"`
	Filename string `json:"filename"`
}

var cephalonDiskDownload = func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	defer r.Body.Close()

	inputBody, err := io.ReadAll(r.Body)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	var input downloadInput
	if err := json.Unmarshal(inputBody, &input); err != nil {
		return http.StatusInternalServerError, err
	}

	myelinHost := os.Getenv("MYELIN_HOST")

	req, err := http.NewRequest("GET", myelinHost+"/download/"+input.MD5, nil)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	myelinResponse, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("myelin do err")
		return http.StatusInternalServerError, err
	}
	defer myelinResponse.Body.Close()

	fmt.Printf("myelin response code is %d\n", myelinResponse.StatusCode)
	//文件不存在
	if myelinResponse.StatusCode == 404 {
		return http.StatusNotFound, nil
	}

	//文件正在向master节点爬取,轮询什么时候爬取完成
	if myelinResponse.StatusCode == 302 {
		return http.StatusFound, nil
	}

	//确认目标文件夹存在
	fullPath := filepath.Join(d.server.Root, input.Target)
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		return http.StatusInternalServerError, err
	}

	// 创建目标文件
	filePath := filepath.Join(fullPath, input.Filename)
	newFile, err := os.Create(filePath)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer newFile.Close()

	//复制文件
	_, err = io.Copy(newFile, myelinResponse.Body)
	if err != nil {
		fmt.Printf("io copy err")
		return http.StatusInternalServerError, err
	}

	return http.StatusCreated, nil
}

var cephalonDiskDownloadProgress = func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
	// 向myelin请求下载进度
	myelinHost := os.Getenv("MYELIN_HOST")
	md5 := r.URL.Query().Get("stream")
	client := sse.NewClient(myelinHost + "/size/" + md5)
	clientCh := make(chan *sse.Event)
	err := client.SubscribeChanRawWithContext(r.Context(), clientCh)
	if err != nil {
		return 0, err
	}

	// 返回读取的数据
	server := sse.New()
	defer server.Close()
	server.CreateStream(md5)
	go func() {
		for {
			select {
			case event := <-clientCh:
				server.TryPublish(md5, &sse.Event{
					Data: event.Data,
				})
			case <-r.Context().Done():
				fmt.Printf("结束")
				return
			}
		}
	}()
	server.ServeHTTP(w, r)

	return 0, nil

}

var cephalonDirInfo = func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
	// 向user_center请求用户信息
	request, err := http.NewRequest("GET", DirInfoURL, nil)
	if err != nil {
		fmt.Printf("http.NewRequest err")
		return 0, err
	}
	query := request.URL.Query()
	query.Add("user_id", os.Getenv("USER_ID"))
	request.URL.RawQuery = query.Encode()

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		fmt.Printf("http.DefaultClient.Do err")
		return http.StatusInternalServerError, err
	}
	defer resp.Body.Close()
	// 读取返回的数据
	body, err := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(body); err != nil {
		return http.StatusInternalServerError, err
	}
	return 0, nil
}

var cephalonUserSpace = func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
	// 向user_center请求用户信息
	request, err := http.NewRequest("GET", DirInfoURL+"/user-space", nil)
	if err != nil {
		fmt.Printf("http.NewRequest err %v", err)
		return 0, err
	}
	query := request.URL.Query()
	query.Add("user_id", os.Getenv("USER_ID"))
	request.URL.RawQuery = query.Encode()

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		fmt.Printf("http.DefaultClient.Do err")
		return http.StatusInternalServerError, err
	}
	defer resp.Body.Close()
	// 读取返回的数据
	body, err := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(body); err != nil {
		return http.StatusInternalServerError, err
	}
	return 0, nil
}
