package logger

import "log"

type Logger struct{ std *log.Logger }

func New() *Logger                                  { return &Logger{std: log.Default()} }
func (l *Logger) Std() *log.Logger                  { return l.std }
func (l *Logger) Printf(format string, args ...any) { l.std.Printf(format, args...) }
