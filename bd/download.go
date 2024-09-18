package bd

import (
	"bufio"
	"context"
	"errors"
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
		var stopChan, restartChan, cancelChan chan struct{}
		if info, ok := DownloadingMap.m[path+outputFilename]; ok {
			info.IsStop = false
			info.IsErr = false
			stopChan = info.StopChan
			restartChan = info.RestarChan
			cancelChan = info.CancelChan
		} else {
			stopChan = make(chan struct{})
			restartChan = make(chan struct{})
			cancelChan = make(chan struct{})
			DownloadingMap.m[path+outputFilename] = &Temple{
				SizeB:      size,
				CurrentB:   0,
				Percentage: 0,
				IsErr:      false,
				IsStop:     false,
				StopChan:   stopChan,
				RestarChan: restartChan,
				CancelChan: cancelChan,
			}
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
		// 创建一个协程池，协程数量
		p, _ := ants.NewPool(2)
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			for {
				select {
				case <-errChan:
					logrus.Error("下载文件失败,暂停下载文件")
					cancel()
				case <-stopChan:
					logrus.Info("暂停下载文件")
					cancel()
				case <-cancelChan:
					logrus.Info("取消下载文件")
					cancel()
					return
				case <-restartChan:
					logrus.Info("继续下载文件")
					go Download(path, accessToken, dlink, outputFilename, size)
					return
				}
			}
		}()
		for i := 0; uint64(i) <= sum; i++ {
			select {
			case <-ctx.Done():
				logrus.Debug("下载被取消")
				return nil
			default:
				wg.Add(1)
				// 确保i的值在循环中不变
				i_ := i
				if uint64(i) == sum {
					p.Submit(func() {
						doRequest(ctx, uri, uint64(i_), 0, path+outputFilename, tmpPath, true, &wg, errChan)
					})
				} else {
					p.Submit(func() {
						doRequest(ctx, uri, uint64(i_), 0, path+outputFilename, tmpPath, false, &wg, errChan)
					})
				}
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
								doRequest(ctx, uri, uint64(i), 0, path+outputFilename, tmpPath, true, &wg, errChan)
							})
						} else {
							p.Submit(func() {
								doRequest(ctx, uri, uint64(i), 0, path+outputFilename, tmpPath, false, &wg, errChan)
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
		if err != nil {
			return err
		}
		// 标记下载完成
		DownloadingMap.Lock()
		DownloadingMap.m[path+outputFilename] = &Temple{
			SizeB:      size,
			CurrentB:   size,
			Percentage: 1,
			IsErr:      false,
		}
		DownloadingMap.Unlock()

		go func() {
			logrus.Info("删除临时文件夹: ", tmpPath)
			err := os.RemoveAll(tmpPath)
			if err != nil {
				logrus.Error(err)
			}
		}()
	default:
		var cancelChan chan struct{}
		var restartChan chan struct{}
		var stopChan chan struct{}
		DownloadingMap.Lock()
		if info, ok := DownloadingMap.m[path+outputFilename]; ok {
			// 重新下载
			info.IsStop = false
			info.IsErr = false
			cancelChan = info.CancelChan
			restartChan = info.RestarChan
			stopChan = info.StopChan
		} else {
			// 新建下载
			cancelChan = make(chan struct{})
			restartChan = make(chan struct{})
			stopChan = make(chan struct{})
			DownloadingMap.m[path+outputFilename] = &Temple{
				SizeB:      size,
				CurrentB:   0,
				Percentage: 0,
				IsErr:      false,
				IsStop:     false,
				IsSmall:    true,
				CancelChan: cancelChan,
				RestarChan: restartChan,
				StopChan:   stopChan,
			}
		}
		DownloadingMap.Unlock()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			for {
				select {
				case <-stopChan:
					cancel()
				case <-cancelChan:
					cancel()
				case <-restartChan:
					go Download(path, accessToken, dlink, outputFilename, size)
					return
				}
			}
		}()
		err := handelDownload(ctx, uri, size, outputFilename, path+outputFilename)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			logrus.Error(err)
			DownloadingMap.Lock()
			info, ok := DownloadingMap.m[path+outputFilename]
			if ok {
				info.IsErr = true
				info.IsStop = true
			}
			DownloadingMap.Unlock()
		}

	}
	return nil
}

func handelDownload(ctx context.Context, uri string, size uint64, outputFilename string, downloadPath string) error {
	var offset int64 = 0
	fileInfo, err := os.Stat(downloadPath)
	if err == nil {
		if fileInfo.Size() == int64(size) {
			logrus.Info("文件: ", outputFilename, "已存在且完整，跳过下载此文件")
			return nil
		} else {
			offset = fileInfo.Size() - 1
		}
	}
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if offset < 0 {
		offset = 0
	}
	if err := downloadFile(ctx, uri, downloadPath, size, offset, 0); err != nil {
		return err
	}
	return nil
}

func downloadFile(ctx context.Context, uri string, saveFilePath string, size uint64, offset, scope int64) error {
	logrus.Info("offset: ", offset, " scope: ", scope, " size: ", size)
	headers := map[string]string{
		"User-Agent": "pan.baidu.com",
	}
	if offset != 0 {
		if scope != 0 {
			headers["Range"] = fmt.Sprintf("bytes=%d-%d", offset, offset+scope-1)
		} else {
			headers["Range"] = fmt.Sprintf("bytes=%d-", offset+1)
		}
	}
	body, statusCode, err := Do2HTTPRequest(ctx, uri, nil, headers)
	if err != nil {
		return err
	}
	logrus.Info("下载文件: ", saveFilePath, " 开始", "statusCode:", statusCode)
	if statusCode != 200 && statusCode != 206 {
		// 重试3次
		for i := 1; i <= 3; i++ {
			body, statusCode, err = Do2HTTPRequest(ctx, uri, nil, headers)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				continue
			}
			if statusCode != 200 && statusCode != 206 {
				continue
			}
			if i == 3 {
				return err
			}
			break
		}
	}
	file, _ := os.OpenFile(saveFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	defer file.Close()
	_, err = io.Copy(file, body)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return nil
		default:
			return err
		}
	}

	return nil
}

func doRequest(ctx context.Context, uri string, index uint64, restart int, downloadPath string, tmpPath string, isEnd bool, wg *sync.WaitGroup, errChan chan struct{}) {
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
	if restart > 10 {
		logrus.Error("下载文件失败，重试次数过多")
		DownloadingMap.Lock()
		info, ok := DownloadingMap.m[downloadPath]
		if ok {
			info.IsErr = true
			info.IsStop = true
		}
		DownloadingMap.Unlock()
		errChan <- struct{}{}
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

	body, statusCode, err := Do2HTTPRequest(ctx, uri, nil, headers)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			logrus.Debug("下载文件失败，下载被取消")
			return
		default:
			logrus.Error(err)

			logrus.Info("开始重新下载文件,下载编号: ", index, " 重载次数: ", restart)
			if restart < 3 {
				time.Sleep(2 * time.Duration(restart) * time.Second)
			} else {
				time.Sleep(2 * time.Duration(restart) * time.Second)
			}
			go doRequest(ctx, uri, index, restart+1, downloadPath, tmpPath, isEnd, wg, errChan)
			return
		}
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
		go doRequest(ctx, uri, index, restart+1, downloadPath, tmpPath, isEnd, wg, errChan)
		return
	}
	// 下载数据输出到名“outputFilename”的文件
	file, err := os.OpenFile(tp, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		logrus.Error(err)
		go doRequest(ctx, uri, index, restart+1, downloadPath, tmpPath, isEnd, wg, errChan)
		return
	}
	n, err := io.Copy(file, body)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			logrus.Debug("下载被取消")
			return
		default:
			go doRequest(ctx, uri, index, restart+1, downloadPath, tmpPath, isEnd, wg, errChan)
			logrus.Error(err)
			return
		}
	}
	if !isEnd {
		if n != 100*MB {
			go doRequest(ctx, uri, index, restart+1, downloadPath, tmpPath, isEnd, wg, errChan)
			logrus.Error("下载文件失败，文件大小不对")
			return
		}
	}
	if wg != nil {
		wg.Done()
	}
}
