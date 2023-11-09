package bd

import (
	"context"
	"encoding/json"
	"errors"
	openapi "github.com/filebrowser/filebrowser/v2/bd/openxpanapi"
	"github.com/sirupsen/logrus"
	"os"
)

// LoginCode 百度用户通过授权码进行登录
type LoginCode struct {
	Code string `json:"code"`
}

var (
	configuration  *openapi.Configuration
	apiClient      *openapi.APIClient
	DownloadingMap map[string]*Temple
)

type Temple struct {
	Size       int     `json:"size"`
	Current    int     `json:"current"`
	Percentage float64 `json:"percentage"`
}

func init() {
	configuration = openapi.NewConfiguration()
	apiClient = openapi.NewAPIClient(configuration)
	DownloadingMap = make(map[string]*Temple)
}
func (code LoginCode) VerifyCode() (string, error) {
	ctx := context.Background()

	resp, _, err := apiClient.AuthApi.OauthTokenCode2token(ctx).Code(code.Code).ClientId("zNBhtXeLhZDRoxMI6trDohpVREC5AEFP").ClientSecret("ZllR6fnf7T7r9qtFpismGmmQ4k4SZ3Ao").RedirectUri("oob").Execute()
	if err != nil {
		logrus.Error(err)
		return "", err
	}
	logrus.Info(resp.AccessToken)
	return resp.GetAccessToken(), nil

}

type GetUserInfo struct {
	AccessToken string `json:"access_token"`
}

type GetUserInfoResp struct {
	HasUsed   *int64  `json:"has_used"`
	TotalCap  *int64  `json:"total_cap"`
	IsVip     *int32  `json:"is_vip"`
	BaiduName *string `json:"baidu_name"`
	Expire    *bool   `json:"expire"`
	Free      *int64  `json:"free"`
}

func (req GetUserInfo) GetUserInfo() (GetUserInfoResp, error) {
	infoResponse, _, _ := apiClient.UserinfoApi.Xpannasuinfo(context.Background()).AccessToken(req.AccessToken).Execute()
	var index openapi.Uinforesponse
	if infoResponse == index || *infoResponse.Errmsg != "succ" {
		logrus.Error()
		return GetUserInfoResp{}, errors.New("fail")
	}

	execute, _, err := apiClient.UserinfoApi.Apiquota(context.Background()).AccessToken(req.AccessToken).Execute()
	if err != nil {
		return GetUserInfoResp{}, err
	}
	resp := GetUserInfoResp{
		HasUsed:   execute.Used,
		TotalCap:  execute.Total,
		IsVip:     infoResponse.VipType,
		BaiduName: infoResponse.BaiduName,
		Expire:    execute.Expire,
		Free:      execute.Free,
	}
	logrus.Info(resp)
	return resp, nil
}

type ShowDirInfoReq struct {
	Path        string `json:"path"`
	AccessToken string `json:"access_token"`
}

func (req ShowDirInfoReq) ShowDirInfo() (*string, error) {
	response, _, err := apiClient.FileinfoApi.Xpanfilelist(context.Background()).AccessToken(req.AccessToken).Dir(req.Path).Execute()
	if err != nil {
		return nil, err
	}
	return &response, nil
}

type DownloadInfoReq struct {
	IsDir       bool   `json:"is_dir"`
	Path        string `json:"path"`
	TargetPath  string `json:"target_path"` // 目标下载路径，请以
	FsID        uint64 `json:"fs_id"`
	AccessToken string `json:"access_token"`
}

func (req DownloadInfoReq) Download() error {
	switch req.IsDir {
	case true:
		response, _, err := apiClient.MultimediafileApi.Xpanfilelistall(context.Background()).AccessToken(req.AccessToken).Recursion(1).Path(req.Path).Execute()
		if err != nil {
			logrus.Error(err)
			return err
		}
		var readFileListRespBody struct {
			Errno int `json:"errno"`
			List  []struct {
				FileName string `json:"server_filename"`
				FsID     int64  `json:"fs_id"`
			} `json:"list"`
		}
		err = json.Unmarshal([]byte(response), &readFileListRespBody)
		if err != nil {
			logrus.Error(err)
			return err
		}
		var fsIDs = make([]uint64, 0)
		for _, fileInfo := range readFileListRespBody.List {
			fsIDs = append(fsIDs, uint64(fileInfo.FsID))
		}
		metasArg := NewFileMetasArg(fsIDs, "./")
		metas, err := FileMetas(req.AccessToken, metasArg)
		if err != nil {
			logrus.Error(err)
			return err
		}
		logrus.Info("file num: ", len(metas.List))

		for _, meta := range metas.List {
			if meta.Isdir == 0 {
				path := req.TargetPath + meta.Path[:len(meta.Path)-len(meta.Filename)]
				err := os.MkdirAll(path, 0777)
				if err != nil {
					return err
				}
				go func() {
					err := Download(path, req.AccessToken, meta.Dlink, meta.Filename, meta.Size)
					if err != nil {
						logrus.Error(err)
					}
				}()
				if err != nil {
					logrus.Error(err)
					return err
				}
			}
		}
	case false:
		fsIDs := []uint64{req.FsID}
		metasArg := NewFileMetasArg(fsIDs, "./")
		metas, err := FileMetas(req.AccessToken, metasArg)
		if err != nil {
			logrus.Error(err)
			return err
		}
		err = os.MkdirAll(req.TargetPath+"/", 0777)
		if err != nil {
			return err
		}
		for _, meta := range metas.List {
			go func() {
				err := Download(req.TargetPath+"/", req.AccessToken, meta.Dlink, meta.Filename, meta.Size)
				if err != nil {
					logrus.Error(err)
				}
			}()

			if err != nil {
				logrus.Error(err)
				return err
			}
		}
	}
	return nil
}

type DownloadProgressReq struct {
	FileName string `json:"file_name"`
}

func (req DownloadProgressReq) GetDownloadProgress() (map[string]*Temple, error) {
	for _, info := range DownloadingMap {
		info.Percentage = float64(info.Current) / float64(info.Size)
	}

	return DownloadingMap, nil
}
