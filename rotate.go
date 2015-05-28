package gotools

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type RotateLogger struct {
	sync.RWMutex
	maxSuffix int
	absPath   string
	prefix    string
	flag      int
	fp        *os.File
	logger    *log.Logger
	rCond     func() bool
	goRotate  bool
	lines     int
	timeout   time.Time
}

func NewRotateLogger(path string, prefix string, flag int, backup int) *RotateLogger {
	if backup < 0 {
		backup = 0
	}
	var absPath string
	var err error
	absPath, err = filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	var fp *os.File
	fp, err = os.OpenFile(absPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	return &RotateLogger{
		maxSuffix: backup,
		absPath:   absPath,
		prefix:    prefix,
		flag:      flag,
		fp:        fp,
		logger:    log.New(fp, prefix, flag),
		rCond:     func() bool { return false },
		goRotate:  false,
		lines:     0,
		timeout:   time.Now(),
	}
}

func (rlogger *RotateLogger) rotateRoutine() {
	for {
		rlogger.Lock()
		if !rlogger.rCond() {
			rlogger.Unlock()
			time.Sleep(time.Second)
			continue
		}
		if rlogger.maxSuffix == 0 {
			rlogger.fp.Truncate(0)
			rlogger.Unlock()
			time.Sleep(time.Second)
			continue
		}
		flist := make([]string, 0, rlogger.maxSuffix+1)
		flist = append(flist, rlogger.absPath)
		for i := 1; i <= rlogger.maxSuffix; i++ {
			flist = append(flist, fmt.Sprintf("%s.%d", rlogger.absPath, i))
		}
		rlogger.fp.Close()
		rlogger.lines = 0

		var err error
		for i := len(flist) - 1; i > 0; i-- {
			if _, err = os.Stat(flist[i-1]); err != nil {
				// file not exist
				continue
			}
			if err = os.Rename(flist[i-1], flist[i]); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
		rlogger.fp, _ = os.OpenFile(rlogger.absPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		rlogger.logger = log.New(rlogger.fp, rlogger.prefix, rlogger.flag)
		rlogger.Unlock()
		time.Sleep(time.Second)
	}
}

func (rlogger *RotateLogger) SetRotateCond(f func() bool) {
	rlogger.Lock()
	defer rlogger.Unlock()
	rlogger.rCond = f
	if !rlogger.goRotate {
		rlogger.goRotate = true
		go rlogger.rotateRoutine()
	}
}

func (rlogger *RotateLogger) SetLineRotate(line int) {
	if line <= 0 {
		return
	}
	rlogger.SetRotateCond(func() bool {
		if rlogger.lines >= line {
			return true
		}
		return false
	})
}

func (rlogger *RotateLogger) SetTimeRotate(t time.Duration) {
	rlogger.timeout = time.Now().Add(t)
	rlogger.SetRotateCond(func() bool {
		if time.Now().After(rlogger.timeout) {
			rlogger.timeout = time.Now().Add(t)
			return true
		}
		return false
	})
}

func (rlogger *RotateLogger) Println(arg ...interface{}) {
	rlogger.RLock()
	rlogger.lines++
	rlogger.logger.Println(arg...)
	rlogger.RUnlock()
}
