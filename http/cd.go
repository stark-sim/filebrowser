package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type downloadInput struct {
	MD5      string `json:"md5"`
	Target   string `json:"target"`
	Filename string `json:"filename"`
}

var cephalonDiskDownload = func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	defer r.Body.Close()
	inputBody, _ := io.ReadAll(r.Body)
	var input downloadInput
	if err := json.Unmarshal(inputBody, &input); err != nil {
		return http.StatusInternalServerError, err
	}

	// 从环境变量 MYELIN_HOST 读取下载文件的节点地址
	myelinHost := os.Getenv("MYELIN_HOST")

	// 调用接口进行下载
	req, _ := http.NewRequest("GET", myelinHost+"/download/"+input.MD5, nil)
	myelinResponse, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("myelin do err")
		return http.StatusInternalServerError, err
	}
	fmt.Printf("myelin response code is  %d\n", myelinResponse.StatusCode)
	// body 就是文件流，直接下载
	defer myelinResponse.Body.Close()
	// 准备好新文件坑位
	newFile, err := os.Create(input.Filename)
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
