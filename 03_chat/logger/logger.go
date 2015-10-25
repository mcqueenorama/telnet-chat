package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/op/go-logging"
)

// logger used throughout
// defaults to stdout
// setup in setupLoggingOrDie

type Logger struct {
	*logging.Logger
}

var Log *logging.Logger

func Info(format string, args ...interface{}) {
	Log.Info(format, args...)
}

func Error(format string, args ...interface{}) {
	Log.Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	Log.Fatalf(format, args...)
}

// setup logging properly or die
// logs are not open yet so write for Std*
// func SetupLoggingOrDie(logFile string) *logging.Logger {
func MustSetupLogging(logFile string) *Logger {

	//default log to stdout
	var logHandle io.WriteCloser = os.Stdout

	var err error

	if logFile != "" {

		if logHandle, err = os.OpenFile(logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666); err != nil {
			fmt.Fprintf(os.Stderr, "Can't open log:%s:err:%v:\n", logFile, err)
			os.Exit(1)
		}

		fmt.Printf("Logging to:logFile:%s:\n", logFile)

	} else {
		fmt.Printf("No logfile specified - going with stdout\n")
	}

	Log, err = logging.GetLogger("chatLog")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't start logger:%s:err:%v:\n", logFile, err)
		os.Exit(1)
	}

	backend1 := logging.NewLogBackend(logHandle, "", 0)
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.INFO, "")
	logging.SetBackend(backend1Leveled)

	return &Logger{Log}

}
