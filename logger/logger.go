package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
)

type Level int

const (
	ErrorLevel Level = 0
	WarnLevel  Level = 1
	InfoLevel  Level = 2
	DebugLevel Level = 3
)

var logLevelFlags = [...]string{"[ERROR]", "[WARN]", "[INFO]", "[DEBUG]"}

type Logger struct {
	out        io.WriteCloser
	level      Level
	logger     *log.Logger
	requestID  string
	callerSkip int
}

var logFlags = log.Ldate | log.Ltime | log.Lmicroseconds

func new(path string, logLevel Level) Logger {
	if path != "" {
		out, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic("failed to open log file" + path)
		}
		return Logger{
			out:    out,
			level:  logLevel,
			logger: log.New(out, "", logFlags),
		}
	}
	out := os.Stdout
	return Logger{
		out:    out,
		level:  logLevel,
		logger: log.New(out, "", logFlags),
	}
}

func (l *Logger) SetRequestID(requestID string) {
	l.requestID = requestID
}

func (l Logger) Close() error {
	return l.out.Close()
}

/**
 * str1:caller
 */
func (l Logger) println(caller string, logLevel Level, args []interface{}) {

	a := make([]interface{}, 0, 10)
	/*//终端颜色
	b := 0
	f := 32
	h := 0
	if logLevel == InfoLevel {
		f = 37
	}
	if logLevel == ErrorLevel {
		h = 1
		b = 41
		f = 37
	}
	if logLevel == WarnLevel {
		h = 1
		b = 43
		f = 31
	}
	*/
	//a = append(a, fmt.Sprintf("%c[%d;%d;%dm%s%c[0m", 0x1B, h, b, f, logLevelFlags[logLevel], 0x1B))
	a = append(a, logLevelFlags[logLevel])
	a = append(a, l.requestID)
	//a = append(a, fmt.Sprintf("%c[0;0;36m%s%c[0m", 0x1B, caller, 0x1B))
	a = append(a, caller)
	a = append(a, args...)
	// fmt.Println(l.logger)
	l.logger.Println(a...)
}
func (l Logger) Debug(args ...interface{}) {
	if l.level < DebugLevel {
		return
	}
	l.println(getCaller(), DebugLevel, args)
}

func (l Logger) INFO(args ...interface{}) {
	if l.level < InfoLevel {
		return
	}
	l.println(getCaller(), InfoLevel, args)
}

func (l Logger) Warn(args ...interface{}) {
	if l.level < WarnLevel {
		return
	}

	l.println(getCaller(), WarnLevel, args)
}

func (l Logger) Error(args ...interface{}) {
	if l.level < ErrorLevel {
		return
	}
	l.println(getCaller(), ErrorLevel, args)
}

func getCaller() string {
	skip := 3
	pc, fullPath, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}

	pcName := runtime.FuncForPC(pc).Name()
	a := strings.Split(fullPath, "/")
	fn := a[len(a)-1]
	return fmt.Sprintf("%s(%s):%d", pcName, fn, line)
}

var logger Logger

func init() {
	//日志输出路径,为空则输出到终端
	logger = new("./adp.log", InfoLevel)
}
func New(path string, logLevel Level) Logger {
	return new(path, logLevel)
}
func Default(path string, logLevel Level) {
	logger = new(path, logLevel)
}

func Debug(args ...interface{}) {
	// fmt.Println("DEBUG:", logger)
	logger.Debug(args...)
}

func INFO(args ...interface{}) {
	logger.INFO(args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}
