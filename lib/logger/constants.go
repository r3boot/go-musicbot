package logger

const (
	LOG_INFO    string = "I"
	LOG_DEBUG   string = "D"
	LOG_WARNING string = "W"
	LOG_FATAL   string = "F"
)

type Logger struct {
	UseTimestamp bool
	UseVerbose   bool
	UseDebug     bool
}
