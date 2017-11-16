package logger

func NewLogger(timestamp, debug bool) *Logger {
	return &Logger{
		UseTimestamp: timestamp,
		UseDebug:     debug,
	}
}
