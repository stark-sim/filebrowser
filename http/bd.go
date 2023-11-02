package http

import (
	"encoding/json"
	"github.com/filebrowser/filebrowser/v2/bd"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

var bdLogin = func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	all, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	var loginInfo bd.LoginCode
	err := json.Unmarshal(all, &loginInfo)
	if err != nil {
		return renderJSON(w, r, err)
	}
	accessToken, err := loginInfo.VerifyCode()
	if err != nil {
		return renderJSON(w, r, err)
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
		return renderJSON(w, r, err)
	}
	info, err := showDirInfo.ShowDirInfo()
	if err != nil {
		return renderJSON(w, r, err)
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
		return renderJSON(w, r, err)
	}
	logrus.Info(downloadInfo)
	err = downloadInfo.Download()
	if err != nil {
		return renderJSON(w, r, err)
	}
	return renderJSON(w, r, d)
}
