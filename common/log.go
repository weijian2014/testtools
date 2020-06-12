package common

import (
	"io"
	"log"
	"os"
)

var (
	logger *Logger = nil
)

const (
	DebugLevel int = iota
	InfoLevel
	SystemLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

type Logger struct {
	level         int
	rollBackLines int
	lines         int
	filePath      string
	file          *os.File
}

func LoggerInit(lev int, roll int, fullPath string) {
	logger = &Logger{level: lev, rollBackLines: roll, lines: 0, filePath: fullPath, file: nil}
	log.SetFlags(log.Ldate | log.Lmicroseconds)

	var w io.Writer
	if "" == logger.filePath {
		w = os.Stdout
	} else {
		createLogFile()
		// Logger实现了Write方法,可以做为io.Writer接口使用
		w = logger
	}

	log.SetOutput(w)
}

func Debug(fmt string, args ...interface{}) {
	if logger.level <= DebugLevel {
		log.SetPrefix("debug ")
		log.Printf(fmt, args...)
	}
}

func Info(fmt string, args ...interface{}) {
	if logger.level <= InfoLevel {
		log.SetPrefix("info  ")
		log.Printf(fmt, args...)
	}
}

func System(fmt string, args ...interface{}) {
	log.SetPrefix("system ")
	log.Printf(fmt, args...)
}

func Warn(fmt string, args ...interface{}) {
	if logger.level <= WarnLevel {
		log.SetPrefix("warn  ")
		log.Printf(fmt, args...)
	}
}

func Error(fmt string, args ...interface{}) {
	if logger.level <= ErrorLevel {
		log.SetPrefix("error ")
		log.Printf(fmt, args...)
	}
}

func Fatal(fmt string, args ...interface{}) {
	if logger.level <= FatalLevel {
		log.SetPrefix("fatal ")
		log.Fatalf(fmt, args...)
	}
}

func createLogFile() {
	if nil != logger.file {
		logger.file.Close()
	}

	f, err := os.OpenFile(logger.filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if nil != err {
		panic(err.Error())
	}

	logger.file = f
}

func (*Logger) Write(buf []byte) (n int, err error) {
	if "" == logger.filePath {
		return logger.file.Write(buf)
	}

	if logger.lines >= logger.rollBackLines {
		createLogFile()
		logger.lines = 0
	}

	logger.lines++
	return logger.file.Write(buf)
}
