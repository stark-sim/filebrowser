package bd

import (
	"encoding/json"
	"errors"
	"io"
	"net/url"
	//	"icode.baidu.com/baidu/xpan/go-sdk/xpan/utils"
)

func FileMetas(accessToken string, arg *FileMetasArg) (FileMetasReturn, error) {
	ret := FileMetasReturn{}

	protocal := "https"
	host := "pan.baidu.com"
	router := "/rest/2.0/xpan/multimedia?method=filemetas&"
	uri := protocal + "://" + host + router

	params := url.Values{}
	params.Set("access_token", accessToken)
	fsidJs, err := json.Marshal(arg.Fsids)
	if err != nil {
		return ret, err
	}
	params.Set("fsids", string(fsidJs))
	params.Set("dlink", "1") // 对于下载，dlink为必选参数，才能拿到dlink下载地址
	params.Set("path", arg.Path)
	params.Set("thumb", "1")
	params.Set("needmedia", "1")
	params.Set("extra", "1")
	uri += params.Encode()

	headers := map[string]string{
		"Host":         host,
		"Content-Type": "application/x-www-form-urlencoded",
	}

	var postBody io.Reader
	body, _, err := DoHTTPRequest(uri, postBody, headers)
	if err != nil {
		return ret, err
	}
	if err = json.Unmarshal([]byte(body), &ret); err != nil {
		return ret, errors.New("unmarshal filemetas body failed,body")
	}
	if ret.Errno != 0 {
		return ret, errors.New("call filemetas failed")
	}
	return ret, nil
}
