package ts3

import (
	"fmt"
	"log"
)

type LogLevel int

const (
	Notice LogLevel = 0
	Debug  LogLevel = 1
	Error  LogLevel = 2
	CmdExc LogLevel = 3
)

const (
	CmdColour    = "\033[1;34m%s\033[0m"
	DebugColour  = "\033[0;36m%s\033[0m"
	ErrorColour  = "\033[1;31m%s\033[0m"
	NoticeColour = "\033[1;36m%s\033[0m"
)

var logging bool = false

// Allow you to turn logging on or off
func LoggingEnabled(enable bool) {
	if enable {
		Log(Debug, "Logging enabled")
	} else {
		Log(Debug, "Logging disabled")
	}

	logging = enable
}

func Log(loglevel LogLevel, format string, a ...interface{}) {
	if logging {
		switch loglevel {
		case Debug:
			log.Printf(DebugColour, fmt.Sprintf("[debug]: "+format, a...))
		case Notice:
			log.Printf(NoticeColour, fmt.Sprintf("[notice]: "+format, a...))
		case Error:
			log.Printf(ErrorColour, fmt.Sprintf("[error]: "+format, a...))
		case CmdExc:
			log.Printf(CmdColour, fmt.Sprintf("[executed]: "+format, a...))
		}
	}
}
