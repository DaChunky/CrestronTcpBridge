package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	LOG_MAIN  = iota
	LOG_FATAL = iota
	LOG_ERROR = iota
	LOG_WARN  = iota
	LOG_INFO  = iota
	LOG_DEBUG = iota
)

var logStrToLevel = map[string]int{
	"LOG_MAIN":  LOG_MAIN,
	"LOG_FATAL": LOG_FATAL,
	"LOG_ERROR": LOG_ERROR,
	"LOG_WARN":  LOG_WARN,
	"LOG_INFO":  LOG_INFO,
	"LOG_DEBUG": LOG_DEBUG,
}

var LogToStdOutInCaseOfError bool = false

func logLevelToString(logLevel int) string {
	ret := ""
	switch logLevel {
	case LOG_MAIN:
		ret = "MAIN"
	case LOG_FATAL:
		ret = "FATAL"
	case LOG_ERROR:
		ret = "ERROR"
	case LOG_WARN:
		ret = "WARNING"
	case LOG_INFO:
		ret = "INFO"
	case LOG_DEBUG:
		ret = "DEBUG"
	}
	return ret
}

// Logger represents a possibiliy to log messages
type Logger interface {
	Log(verbose int, msg string)
	LogFmt(verbose int, msg string, v ...interface{})
	SetLogLevel(newLevel int)
}

type logger struct {
	formatString string
	toStdOut     bool
	logLevel     int
}

func (lg *logger) Log(verbose int, msg string) {
	if verbose > lg.logLevel {
		return
	}
	levelStr := logLevelToString(verbose)
	if lg.toStdOut {
		fmt.Printf("%-12s %s\n", levelStr, msg)
	} else {
		log.Printf("%-12s %s\n", levelStr, msg)
	}
}

func (lg *logger) LogFmt(verbose int, msg string, v ...interface{}) {
	lg.Log(verbose, fmt.Sprintf(msg, v...))
}

func (lg *logger) SetLogLevel(newLevel int) {
	lg.logLevel = newLevel
}

func tryOpenLogSettingsAndDetermineMinLogLevel() int {
	res := LOG_DEBUG
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	basePath := ""
	if path_seperator == "\\" {
		// build on windows. expecting log file settings in programData
		basePath = os.Getenv("ALLUSERSPROFILE")
	} else {
		// build on linux
		basePath = "/etc"
	}
	pathToConfig := filepath.Join(basePath, appName, fmt.Sprintf("%s_log.sett", appName))
	content, err := os.ReadFile(pathToConfig)
	if err == nil {
		contentStr := strings.TrimSuffix(strings.ToUpper(string(content)), "\n")
		tryRes, ok := logStrToLevel[contentStr]
		if ok {
			res = tryRes
		}
	} else {
		fmt.Printf("%s: %s\n", pathToConfig, err)
	}
	return res
}

func CreateAndInitLog(file string, force bool) (Logger, error) {
	ret := new(logger)
	ret.formatString = "%s: %s"
	ret.toStdOut = false
	lgFl, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		ret.toStdOut = LogToStdOutInCaseOfError && force
		if ret.toStdOut {
			return ret, nil
		}
		fmt.Printf("log init failed: %v\n", err)
		return nil, err
	}
	log.SetOutput(lgFl)
	ret.logLevel = tryOpenLogSettingsAndDetermineMinLogLevel()
	return ret, nil
}

var mainLog Logger = nil

func CreateAndInitMainLog() error {
	mainAppCallSplit := strings.Split(os.Args[0], string(os.PathSeparator))
	logName := mainAppCallSplit[len(mainAppCallSplit)-1]
	logFile := fmt.Sprintf("/var/log/%s/%s.log", logName, logName)
	var err error = nil
	mainLog, err = CreateAndInitLog(logFile, false)
	if err != nil {
		logFile = fmt.Sprintf("./%s.log", logName)
		mainLog, err = CreateAndInitLog(logFile, true)
		if err != nil {
			fmt.Printf("main log init failed: %v\n", err)
			return err
		}
	}
	mainLog.Log(LOG_INFO, "---------------------------------------")
	mainLog.Log(LOG_INFO, "application started")
	return nil
}

func checkMainLogInitialized() bool {
	if mainLog == nil {
		err := CreateAndInitMainLog()
		if err != nil {
			return false
		}
	}
	return true
}

func Log(verbose int, msg string) {
	if !checkMainLogInitialized() {
		return
	}
	mainLog.Log(verbose, msg)
}

func LogFmt(verbose int, msg string, v ...interface{}) {
	Log(verbose, fmt.Sprintf(msg, v...))
}
