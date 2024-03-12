package http

import (
	"encoding/json"
	"fmt"
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

	//文件正在向master节点爬取,请稍后尝试
	if myelinResponse.StatusCode == 302 {
		return http.StatusFound, nil
	}

	//确认目标文件夹存在
	if err := os.MkdirAll(input.Target, os.ModePerm); err != nil {
		return http.StatusInternalServerError, err
	}

	// 创建目标文件
	filePath := filepath.Join(input.Target, input.Filename)
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
