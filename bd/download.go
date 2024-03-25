package bd

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
	// "icode.baidu.com/baidu/xpan/go-sdk/xpan/utils"
)

const KB = 1024
const MB = 1024 * KB

var queueChannel = make(chan struct{}, runtime.NumCPU())

func Download(path string, accessToken string, dlink string, outputFilename string, size uint64) error {
	uri := dlink + "&" + "access_token=" + accessToken
	switch {
	case size > 10*MB:
		sum := size / (10 * MB)
		DownloadingMap.Lock()
		DownloadingMap.m[path+outputFilename] = &Temple{
			Size:     int(sum + 1),
			SizeB:    size,
			Current:  0,
			CurrentB: 0,
		}
		DownloadingMap.Unlock()

		var wg sync.WaitGroup
		logrus.Info("下载文件: ", outputFilename, " 共临时文件", sum)
		// 创建临时文件夹
		err := os.MkdirAll(path+"tmp/", os.ModePerm)
		if err != nil {
			return err
		}

		for i := 0; uint64(i) <= sum; i++ {
			wg.Add(1)
			// 确保i的值在循环中不变
			i_ := i
			if uint64(i) == sum {
				go doRequest(uri, uint64(i_), 0, path+outputFilename, path+"tmp/"+outputFilename, true, &wg)
			} else {
				go doRequest(uri, uint64(i_), 0, path+outputFilename, path+"tmp/"+outputFilename, false, &wg)
			}
		}
		wg.Wait()
		file, err := os.OpenFile(path+outputFilename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		defer file.Close()
		// 创建一个带缓冲的写入器，缓冲区大小为10MB
		fileWriter := bufio.NewWriterSize(file, 10*MB)
		if err != nil {
			fmt.Printf("无法写入文件 %s: %v\n", outputFilename, err)
			return err
		}
		for i := 0; uint64(i) <= sum; i++ {
			filename := path + "tmp/" + outputFilename + "-" + strconv.FormatUint(uint64(i), 10)
			tmpFile, err2 := os.Open(filename)
			//使用匿名函数，defer确保关闭文件
			func() {
				defer tmpFile.Close()
				if err2 != nil {
					logrus.Error(err2)
				}
				_, err := fileWriter.ReadFrom(tmpFile)
				if err != nil {
					return
				}
				if err != nil {
					logrus.Error(err)
				}
				/*go func() {
					err := os.Remove(filename)
					if err != nil {
						logrus.Error(err)
					}
				}()*/
			}()
			if err != nil {
				return err
			}
		}
		err = fileWriter.Flush()
		if err != nil {
			return err
		}
	default:
		headers := map[string]string{
			"User-Agent": "pan.baidu.com",
		}

		var postBody io.Reader
		body, statusCode, err := Do2HTTPRequest(uri, postBody, headers)
		if err != nil {
			return err
		}
		if statusCode != 200 {
			return errors.New("download http fail")
		}

		// 下载数据输出到名“outputFilename”的文件
		file, err := os.OpenFile(path+outputFilename, os.O_WRONLY|os.O_CREATE, 0666)
		defer file.Close()
		write := bufio.NewWriter(file)
		_, err = write.ReadFrom(body)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func doRequest(uri string, index uint64, restart int, downloadPath string, tmpPath string, isEnd bool, wg *sync.WaitGroup) {
	fileInfo, err := os.Stat(tmpPath)
	dp := downloadPath + "-" + strconv.FormatUint(index, 10)
	tp := tmpPath + "-" + strconv.FormatUint(index, 10)
	if err == nil && fileInfo.Size() == int64(10*MB) {
		logrus.Info("切片文件:", dp, "已存在且完整，跳过下载此切片文件")
		DownloadingMap.Lock()
		DownloadingMap.m[downloadPath].Current++
		DownloadingMap.m[downloadPath].CurrentB += 10 * MB
		DownloadingMap.Unlock()
		wg.Done()
		return
	}
	queueChannel <- struct{}{}
	time.Sleep(time.Duration(len(queueChannel)) * 1500 * time.Millisecond)
	headers := map[string]string{
		"User-Agent": "pan.baidu.com",
	}
	if isEnd {
		headers["Range"] = "bytes=" + strconv.FormatUint(10*MB*index, 10) + "-"
	} else {
		headers["Range"] = "bytes=" + strconv.FormatUint(10*MB*index, 10) + "-" + strconv.FormatUint(10*MB*(index+1)-1, 10)
	}

	body, statusCode, err := Do2HTTPRequest(uri, nil, headers)
	defer body.Close()
	if err != nil {
		logrus.Error(err)
		logrus.Info("开始重新下载文件,下载编号: ", index, " 重载次数: ", restart)
		if restart < 3 {
			time.Sleep(2 * time.Duration(restart) * time.Second)
		} else {
			time.Sleep(2 * time.Duration(restart) * time.Second)
		}
		go doRequest(uri, index, restart+1, downloadPath, tmpPath, isEnd, wg)
		return
	}

	if statusCode != 200 && statusCode != 206 {
		logrus.Error(err)

		logrus.Info("开始重新下载文件,下载编号: ", index, " 重载次数: ", restart)
		if restart < 3 {
			time.Sleep(2 * time.Duration(restart) * time.Second)
		} else {
			time.Sleep(2 * time.Duration(restart) * time.Second)
		}
		go doRequest(uri, index, restart+1, downloadPath, tmpPath, isEnd, wg)
		return
	}
	// 下载数据输出到名“outputFilename”的文件
	file, err := os.OpenFile(tp, os.O_WRONLY|os.O_CREATE, 0666)
	defer file.Close()
	write := bufio.NewWriterSize(file, 10*MB)
	_, err = write.ReadFrom(body)
	if err != nil {
		return
	}
	err = write.Flush()
	if err != nil {
		return
	}
	<-queueChannel
	DownloadingMap.Lock()
	DownloadingMap.m[downloadPath].Current++
	DownloadingMap.m[downloadPath].CurrentB += 10 * MB
	DownloadingMap.Unlock()
	wg.Done()
}
