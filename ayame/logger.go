package ayame

// 参考: Golang でログを吐くコツ https://www.kaoriya.net/blog/2018/12/16/

import (
	"log"
	"os"
	"sync"
)

var (
	logger = log.New(os.Stderr, "", log.LstdFlags)
	logMu  sync.Mutex
)

// SetLogger は ayame パッケージ内で出力される *log.Logger を任意のものに設定します。
func SetLogger(l *log.Logger) {
	if l == nil {
		l = log.New(os.Stderr, "", log.LstdFlags)
	}
	logMu.Lock()
	logger = l
	logMu.Unlock()
}

func logf(format string, v ...interface{}) {
	logMu.Lock()
	logger.Printf(format, v...)
	logMu.Unlock()
}
