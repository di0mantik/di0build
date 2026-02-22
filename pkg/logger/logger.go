package logger

import (
	"log"
	"os"
)

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	yellow = "\033[33m"
	green  = "\033[32m"
	blue   = "\033[34m"
)

var std = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds)

func Info(format string, args ...any) {
	std.Printf(green+"[INFO] "+reset+format+"\n", args...)
}

func Warn(format string, args ...any) {
	std.Printf(yellow+"[WARN] "+reset+format+"\n", args...)
}

func Error(format string, args ...any) {
	std.Printf(red+"[ERROR] "+reset+format+"\n", args...)
}

func Fatal(format string, args ...any) {
	std.Printf(red+"[FATAL] "+reset+format+"\n", args...)
	os.Exit(1)
}
