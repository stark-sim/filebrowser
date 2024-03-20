package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/filebrowser/filebrowser/v2/bd"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
)

var bdUserInfo = func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	all, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	var userInfoReq bd.GetUserInfo
	err := json.Unmarshal(all, &userInfoReq)
	if err != nil {
		logrus.Error(err)
		return http.StatusInternalServerError, nil
	}
	resp, err := userInfoReq.GetUserInfo()
	if err != nil {
		logrus.Error(err)
		return http.StatusUnauthorized, nil
	}
	logrus.Info(resp)
	marshal, err := json.Marshal(resp)
	if err != nil {
		logrus.Error(err)
		return http.StatusInternalServerError, nil
	}
	if _, err := w.Write(marshal); err != nil {
		return http.StatusInternalServerError, nil
	}
	return 0, nil
}
var bdLogin = func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	all, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	var loginInfo bd.LoginCode
	err := json.Unmarshal(all, &loginInfo)
	if err != nil {
		return http.StatusInternalServerError, nil
	}
	accessToken, err := loginInfo.VerifyCode()
	if err != nil {
		return http.StatusUnauthorized, nil
	}
	if accessToken == "" {
		return http.StatusUnauthorized, nil
	}
	return renderJSON(w, r, map[string]string{
		"access_token": accessToken,
	})
}

var bdShowDirInfo = func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	all, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	var showDirInfo bd.ShowDirInfoReq
	err := json.Unmarshal(all, &showDirInfo)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	info, err := showDirInfo.ShowDirInfo()
	if err != nil {
		if errors.Is(err, bd.InvalidAuth) {
			return http.StatusUnauthorized, err
		}
		return http.StatusInternalServerError, err

	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write([]byte(*info)); err != nil {
		return http.StatusInternalServerError, err
	}
	return 0, nil
}

var bdDownLoad = func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	all, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	var downloadInfo bd.DownloadInfoReq
	err := json.Unmarshal(all, &downloadInfo)
	if err != nil {
		logrus.Error(err)
		return http.StatusInternalServerError, err
	}
	logrus.Info(downloadInfo)
	_, err = downloadInfo.ShowDirInfo()
	if err != nil {
		if errors.Is(err, bd.InvalidAuth) {
			return http.StatusUnauthorized, err
		}
		return http.StatusInternalServerError, err
	}
	err = downloadInfo.Download(d.server.Root)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return renderJSON(w, r, d)
}

var bdDownloadProgress = func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	all, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	var downloadProgress bd.DownloadProgressReq
	err := json.Unmarshal(all, &downloadProgress)
	if err != nil {
		logrus.Error(err)
		return renderJSON(w, r, err)
	}
	logrus.Info(string(all))
	percentage, err := downloadProgress.GetDownloadProgress()
	if err != nil {
		return 0, err
	}
	return renderJSON(w, r, percentage)
}

var bdRefreshAccessToken = func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	// 向user_center请求刷新用户百度token
	request, err := http.NewRequest("PUT", os.Getenv("USER_CENTER_HOST")+"/v1/cloud-files/baidu-access-token/refresh", nil)
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

var bdGetAccessToken = func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	// 向user_center请求用户信息
	request, err := http.NewRequest("GET", os.Getenv("USER_CENTER_HOST")+"/v1/cloud-files/baidu-access-token", nil)
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
