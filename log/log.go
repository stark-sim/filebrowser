package log

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(formatter(true))
	logrus.SetReportCaller(true)
	logrus.SetLevel(logrus.DebugLevel)

	logrus.Debugln("[INIT] init log success")
}

func formatter(isConsole bool) *nested.Formatter {
	fmtter := &nested.Formatter{
		HideKeys:        false,
		TimestampFormat: "2006-01-02 15:04:05",
		CallerFirst:     true,
		ShowFullLevel:   true,
		CustomCallerFormatter: func(frame *runtime.Frame) string {
			funcInfo := runtime.FuncForPC(frame.PC)
			if funcInfo == nil {
				return "error during runtime.FuncForPC"
			}
			fullPath, line := funcInfo.FileLine(frame.PC)
			fncSlice := strings.Split(funcInfo.Name(), ".")
			fncName := fncSlice[len(fncSlice)-1]
			return fmt.Sprintf(" [%v]-[%v]-[%v]", filepath.Base(fullPath), fncName, line)
		},
	}
	if isConsole {
		fmtter.NoColors = false
	} else {
		fmtter.NoColors = true
	}
	return fmtter
}
