package utils

import (
	"github.com/juju/ratelimit"
	"io"
	"net/http"
	"os"
)

// 基于开源库 reatlimit的令牌桶算法实现的限流器
const (
	CEPHALON_CORE_MODE      = "CEPHALON_CORE_MODE"
	CEPHALON_CORE_MODE_TEST = "test"
	CEPHALON_CORE_MODE_PROD = "prod"
)

// 全局限流器 只限速本地上传
const downLoadLimit3Mb = 3 * 1024 * 1024 // 3Mb/s

var uploadBucket = ratelimit.NewBucketWithRate(downLoadLimit3Mb, downLoadLimit3Mb)   //全局的一个 不论同时上传多少速度总和都是3Mb/s
var downloadBucket = ratelimit.NewBucketWithRate(downLoadLimit3Mb, downLoadLimit3Mb) //全局的一个 不论同时下载多少速度总和都是3Mb/s

var _ http.ResponseWriter = (*downloadBucketWriter)(nil) // 强制检查是否实现了http.ResponseWriter接口
var _ io.Writer = (*downloadBucketWriter)(nil)           // 强制检查是否实现了io.Writer接口

// 下载限流器  //用户从服务器下载文件时，限制下载速度
type downloadBucketWriter struct {
	w      http.ResponseWriter
	bucket *ratelimit.Bucket
}

func (d *downloadBucketWriter) Header() http.Header {
	return d.w.Header()
}
func (d *downloadBucketWriter) Write(bt []byte) (int, error) {
	mode := os.Getenv(CEPHALON_CORE_MODE)
	if mode == CEPHALON_CORE_MODE_TEST {
		return d.w.Write(bt) //test 环境不限速
	} else {
		return ratelimit.Writer(d.w, d.bucket).Write(bt)
	}

}
func (d *downloadBucketWriter) WriteHeader(statusCode int) {
	d.w.WriteHeader(statusCode)
}

func NewDownloadBucketWriter(w http.ResponseWriter) *downloadBucketWriter {
	return &downloadBucketWriter{
		w:      w,
		bucket: downloadBucket,
	}
}

// 上传限流器 //用户上传文件时，限制上传速度
func NewUploadBucketWriter(w io.Writer) io.Writer {
	mode := os.Getenv(CEPHALON_CORE_MODE)
	if mode == CEPHALON_CORE_MODE_TEST {
		return w //test 环境不限速
	} else {
		return ratelimit.Writer(w, uploadBucket)
	}
}
