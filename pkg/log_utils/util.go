package log_utils

import (
	"log"
	"os"
)

//
// @Author yfy2001
// @Date 2024/12/17 15 08
//

// 定义颜色代码
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
)

var (
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger
)

// init 函数在包被导入时自动调用
func init() {
	Info = createLogger("INFO  ", colorGreen)
	Warn = createLogger("WARN  ", colorYellow)
	Error = createLogger("ERROR ", colorRed)
}

// 创建日志记录器
func createLogger(prefix string, color string) *log.Logger {
	return log.New(os.Stdout, color+prefix+colorReset, log.Lshortfile|log.LstdFlags)
}
