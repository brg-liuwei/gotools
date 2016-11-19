package gotools

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"time"
)

type printer interface {
	Println(...interface{})
	Printf(...interface{})
}

type Estimate struct {
	name   string
	last   time.Time
	logger printer
	buf    *bytes.Buffer
}

var defaultLogger *log.Logger

func init() {
	defaultLogger = log.New(os.Stderr, "", log.Ldate|log.Ltime)
}

func NewEstimate(name string) *Estimate {
	return &Estimate{
		name:   name,
		logger: defaultLogger,
		buf:    bytes.NewBufferString("\n    "),
	}
}

func (es *Estimate) SetLogger(logger *log.Logger) {
	es.logger = logger
}

func (es *Estimate) Add(s string) {
	var since time.Duration
	var info string

	if !es.last.IsZero() {
		since = time.Since(es.last)
		info += fmt.Sprintf(" duration: [%v]\n    ", since)
	}
	now := time.Now()
	es.last = now

	info += fmt.Sprintf("[%v] [%s]", now, s)

	es.buf.Grow(len(info))
	es.buf.WriteString(info)
}

func (es *Estimate) ToString() string {
	return es.buf.String()
}

func (es *Estimate) Write() {
	es.logger.Println(es.buf.String())
}
