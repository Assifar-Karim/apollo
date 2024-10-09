package utils

import (
	"fmt"
	"log"
	"os"
	"slices"
	"sync"
)

var lock = &sync.Mutex{}

type Logger struct {
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	traceLogger *log.Logger
}

var loggerInstance *Logger

func GetLogger() *Logger {
	if loggerInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		flags := log.Ldate | log.Ltime | log.Lmsgprefix

		args := os.Args[1:]
		var traceLogger *log.Logger = nil
		if slices.Contains(args, "--trace") {
			traceLogger = log.New(os.Stdout, "\033[32mTRACE: \033[0m", flags)
		}

		loggerInstance = &Logger{
			infoLogger:  log.New(os.Stdout, "\033[35mINFO: \033[0m", flags),
			warnLogger:  log.New(os.Stdout, "\033[33mWARN: \033[0m", flags),
			errorLogger: log.New(os.Stderr, "\033[31mERROR: \033[0m", flags),
			traceLogger: traceLogger,
		}
	}
	return loggerInstance
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.infoLogger.Printf(format+"\n", v...)
}

func (l *Logger) Warn(format string, v ...interface{}) {
	l.warnLogger.Printf(format+"\n", v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.errorLogger.Printf(format, v...)
}

func (l *Logger) Trace(format string, v ...interface{}) {
	if l.traceLogger != nil {
		l.traceLogger.Printf(format, v...)
	}
}

func (l *Logger) PrintBanner() {
	fmt.Println("    ___                   __     __        ")
	fmt.Println("   /   |   ____   ____   / /    / /   ____ ")
	fmt.Println("  / /| |  / __ \\ / __ \\ / /    / /   / __ \\")
	fmt.Println(" / ___ | / /_/ // /_/ // /___ / /___/ /_/ /")
	fmt.Println("/_/  |_|/ .___/ \\____//_____//_____/\\____/ ")
	fmt.Println("       /_/                                 ")
}
