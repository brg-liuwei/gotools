package gotools

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
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
	lines     int32
	timeout   time.Time
}

func NewRotateLogger(path string, prefix string, flag int, backup int) (*RotateLogger, error) {
	if backup < 0 {
		backup = 0
	}
	var absPath string
	var err error
	absPath, err = filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	var fp *os.File
	fp, err = os.OpenFile(absPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
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
	}, nil
}

func (rlogger *RotateLogger) forceRotateWithLock() {
	if stat, err := os.Stat(rlogger.absPath); err == nil {
		// do not rotate empty log file
		if stat.Size() == 0 {
			return
		}
	}
	if rlogger.maxSuffix == 0 {
		rlogger.fp.Truncate(0)
		return
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
}

func (rlogger *RotateLogger) rotate() {
	rlogger.Lock()
	defer rlogger.Unlock()
	if !rlogger.rCond() {
		return
	}
	rlogger.forceRotateWithLock()
}

func (rlogger *RotateLogger) rotateRoutine() {
	rlogger.Lock()
	rlogger.forceRotateWithLock()
	rlogger.Unlock()
	for {
		time.Sleep(time.Second)
		rlogger.rotate()
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
		if rlogger.lines >= int32(line) {
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
	defer rlogger.RUnlock()
	atomic.AddInt32(&rlogger.lines, 1)
	rlogger.logger.Println(arg...)
}

func (rlogger *RotateLogger) Printf(fmt string, arg ...interface{}) {
	rlogger.RLock()
	defer rlogger.RUnlock()
	atomic.AddInt32(&rlogger.lines, 1)
	rlogger.logger.Printf(fmt, arg...)
}
