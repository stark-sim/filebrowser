package bd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/panjf2000/ants"
	"github.com/sirupsen/logrus"
	// "icode.baidu.com/baidu/xpan/go-sdk/xpan/utils"
)

const (
	KB = 1024
	MB = 1024 * KB
)

func Download(path string, accessToken string, dlink string, outputFilename string, size uint64) error {
	uri := dlink + "&" + "access_token=" + accessToken
	switch {
	case size > 100*MB:
		sum := size / (100 * MB)
		DownloadingMap.Lock()
		DownloadingMap.m[path+outputFilename] = &Temple{
			SizeB:      size,
			CurrentB:   0,
			Percentage: 0,
			IsErr:      false,
		}
		DownloadingMap.Unlock()

		var wg sync.WaitGroup
		logrus.Info("下载文件: ", outputFilename, " 共临时文件", sum)
		// 创建临时文件夹
		tmpPath := path + "." + outputFilename + "_tmp/"
		err := os.MkdirAll(tmpPath, os.ModePerm)
		if err != nil {
			return err
		}
		errChan := make(chan struct{})
		isErr := false
		// 创建一个协程池，协程数量
		p, _ := ants.NewPool(2)
		for i := 0; uint64(i) <= sum; i++ {
			select {
			case <-errChan:
				logrus.Error("下载文件失败")
			default:
				wg.Add(1)
				// 确保i的值在循环中不变
				i_ := i
				if uint64(i) == sum {
					p.Submit(func() {
						doRequest(uri, uint64(i_), 0, path+outputFilename, tmpPath, true, &wg, errChan)
					})
				} else {
					p.Submit(func() {
						doRequest(uri, uint64(i_), 0, path+outputFilename, tmpPath, false, &wg, errChan)
					})
				}
			}
			if isErr {
				break
			}
		}
		wg.Wait()
		file, err := os.OpenFile(path+outputFilename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			fmt.Printf("无法写入文件 %s: %v\n", outputFilename, err)
			return err
		}
		defer file.Close()
		// 创建一个带缓冲的写入器，缓冲区大小为10MB
		fileWriter := bufio.NewWriterSize(file, 100*MB)
		for i := 0; uint64(i) <= sum; i++ {
			func() {
				time1 := time.Now()
				filename := tmpPath + outputFilename + "-" + strconv.FormatUint(uint64(i), 10)
				tmpFile, err2 := os.Open(filename)
				// 使用匿名函数，defer确保关闭文件
				if err2 != nil {
					// 重试3次重新下载
					logrus.Error(err2)
					for j := 1; j <= 3; j++ {
						wg.Add(1)
						if uint64(i) == sum {
							p.Submit(func() {
								doRequest(uri, uint64(i), 0, path+outputFilename, tmpPath, true, &wg, errChan)
							})
						} else {
							p.Submit(func() {
								doRequest(uri, uint64(i), 0, path+outputFilename, tmpPath, false, &wg, errChan)
							})
						}
						wg.Wait()
						tmpFile, err2 = os.Open(filename)
						if err2 == nil {
							break
						}
					}
					// 无法打开文件，重新下载改文件
				}
				defer tmpFile.Close()
				logrus.Info("读取文件: ", filename, " 耗时: ", time.Since(time1))
				_, err := fileWriter.ReadFrom(tmpFile)
				if err != nil {
					// 重试3次
					logrus.Error(err)
					for j := 1; j <= 3; j++ {
						_, err = fileWriter.ReadFrom(tmpFile)
						if err == nil {
							break
						}
						logrus.Error(err)
					}
				}
				logrus.Info("合并文件: ", filename, " 耗时: ", time.Since(time1))
			}()
		}
		err = fileWriter.Flush()
		// 标记下载完成
		DownloadingMap.Lock()
		DownloadingMap.m[path+outputFilename] = &Temple{
			SizeB:      size,
			CurrentB:   size,
			Percentage: 1,
			IsErr:      false,
		}
		DownloadingMap.Unlock()

		if err != nil {
			return err
		}
		go func() {
			logrus.Info("删除临时文件夹: ", tmpPath)
			err := os.RemoveAll(tmpPath)
			if err != nil {
				logrus.Error(err)
			}
		}()
	default:
		headers := map[string]string{
			"User-Agent": "pan.baidu.com",
		}

		var postBody io.Reader
		body, statusCode, err := Do2HTTPRequest(uri, postBody, headers)
		if err != nil {
			return err
		}
		if statusCode != 200 && statusCode != 206 {
			// 重试3次
			for i := 1; i <= 3; i++ {
				body, statusCode, err = Do2HTTPRequest(uri, postBody, headers)
				if err == nil {
					break
				}
				if statusCode != 200 && statusCode != 206 {
					continue
				}
				if i == 3 {
					return err
				}
			}
		}

		// 下载数据输出到名“outputFilename”的文件
		file, _ := os.OpenFile(path+outputFilename, os.O_WRONLY|os.O_CREATE, 0666)
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

func doRequest(uri string, index uint64, restart int, downloadPath string, tmpPath string, isEnd bool, wg *sync.WaitGroup, errChan chan struct{}) {
	dp := downloadPath + "-" + strconv.FormatUint(index, 10)
	filename := filepath.Base(downloadPath)
	tp := tmpPath + filename + "-" + strconv.FormatUint(index, 10)
	fileInfo, err := os.Stat(tp)
	if err == nil && fileInfo.Size() == int64(100*MB) {
		logrus.Info("切片文件:", dp, "已存在且完整，跳过下载此切片文件")
		if wg != nil {
			wg.Done()
		}
		return
	}
	time.Sleep(time.Duration(restart) * 3000 * time.Millisecond)
	headers := map[string]string{
		"User-Agent": "pan.baidu.com",
	}
	if isEnd {
		headers["Range"] = "bytes=" + strconv.FormatUint(100*MB*index, 10) + "-"
	} else {
		headers["Range"] = "bytes=" + strconv.FormatUint(100*MB*index, 10) + "-" + strconv.FormatUint(100*MB*(index+1)-1, 10)
	}

	body, statusCode, err := Do2HTTPRequest(uri, nil, headers)
	if err != nil {
		logrus.Error(err)
		if restart > 10 {
			logrus.Error("下载文件失败，重试次数过多")
			DownloadingMap.Lock()
			_, ok := DownloadingMap.m[downloadPath]
			if ok {
				DownloadingMap.m[downloadPath].IsErr = true
			}
			DownloadingMap.Unlock()
			errChan <- struct{}{}
			return
		}
		logrus.Info("开始重新下载文件,下载编号: ", index, " 重载次数: ", restart)
		if restart < 3 {
			time.Sleep(2 * time.Duration(restart) * time.Second)
		} else {
			time.Sleep(2 * time.Duration(restart) * time.Second)
		}
		go doRequest(uri, index, restart+1, downloadPath, tmpPath, isEnd, wg, errChan)
		return
	}
	defer body.Close()

	if statusCode != 200 && statusCode != 206 {
		logrus.Error("下载文件失败，http状态码: ", statusCode)
		data, _ := io.ReadAll(body)
		logrus.Error(string(data))
		logrus.Info("开始重新下载文件,下载编号: ", index, " 重载次数: ", restart)
		if restart < 3 {
			time.Sleep(2 * time.Duration(restart) * time.Second)
		} else {
			time.Sleep(2 * time.Duration(restart) * time.Second)
		}
		go doRequest(uri, index, restart+1, downloadPath, tmpPath, isEnd, wg, errChan)
		return
	}
	// 下载数据输出到名“outputFilename”的文件
	file, err := os.OpenFile(tp, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		logrus.Error(err)
		go doRequest(uri, index, restart+1, downloadPath, tmpPath, isEnd, wg, errChan)
		return
	}
	n, err := io.Copy(file, body)
	if err != nil {
		go doRequest(uri, index, restart+1, downloadPath, tmpPath, isEnd, wg, errChan)
		logrus.Error(err)
		return
	}
	if !isEnd {
		if n != 100*MB {
			go doRequest(uri, index, restart+1, downloadPath, tmpPath, isEnd, wg, errChan)
			logrus.Error("下载文件失败，文件大小不对")
			return
		}
	}

	if wg != nil {
		wg.Done()
	}
}
