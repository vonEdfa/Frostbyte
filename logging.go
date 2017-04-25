package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"
)

const (
	// LogError level is used for critical errors that could lead to data loss
	// or panic that would not be returned to a calling function.
	LogError int = iota

	// LogWarning level is used for very abnormal events and errors that are
	// also returend to a calling function.
	LogWarning

	// LogInformational level is used for normal non-error activity
	LogInformational

	// LogDebug level is for very detailed non-error activity.  This is
	// very spammy and will impact performance.
	LogDebug
)

// LogLevel is for setting the log level
var LogLevel int

// msglog provides package wide logging consistancy for discordgo
// the format, a...  portion this command follows that of fmt.Printf
//   logLevel   : LogLevel of the message
//   caller : 1 + the number of callers away from the message source
//   format : Printf style message format
//   a ...  : comma seperated list of values to pass
func msglog(logLevel, caller int, format string, a ...interface{}) {

	pc, file, line, _ := runtime.Caller(caller)

	files := strings.Split(file, "/")
	file = files[len(files)-1]

	name := runtime.FuncForPC(pc).Name()
	fns := strings.Split(name, ".")
	name = fns[len(fns)-1]

	msg := fmt.Sprintf(format, a...)

	var level string

	switch logLevel {
	case LogError:
		level = "ERROR"
	case LogWarning:
		level = "WARN"
	case LogInformational:
		level = "INFO"
	case LogDebug:
		level = "DEBUG"
	default:
		level = fmt.Sprintf("%d", logLevel)
	}

	if logLevel == LogDebug {
		log.Printf("[%s] %s:%d:%s() %s\n", level, file, line, name, msg)
	} else {
		log.Printf("[%s] %s\n", level, msg)
	}
}

// helper function that wraps msglog for the Session struct
// This adds a check to insure the message is only logged
// if the session log level is equal or higher than the
// message log level
func veeLog(logLevel int, format string, a ...interface{}) {

	if logLevel > LogLevel {
		return
	}

	msglog(logLevel, 2, format, a...)
}
