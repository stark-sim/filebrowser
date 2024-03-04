package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const DirInfoURL = "https://"

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
	if myelinResponse.StatusCode == 404 {
		return http.StatusNotFound, nil
	}

	//todo 节点正在爬取文件，稍后再试
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
