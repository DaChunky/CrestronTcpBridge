package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dachunky/crestrontcpbridge/pkg/crebrid"
	"github.com/dachunky/crestrontcpbridge/pkg/logging"
)

func main() {
	logging.LogToStdOutInCaseOfError = true
	logging.Log(logging.LOG_MAIN, "[main] start crestron bridge service")
	// setup reaction on os signals
	sigs := make(chan os.Signal, 1)
	// catch all signals
	// signal.Notify(sigs, syscall.SIGINT, syscall.SIGSTOP, syscall.SIGQUIT, syscall.SIGTERM)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	path2Cfg := flag.String("config", "/etc/crebrid/crebrid.conf", "app config file")
	setts, err := crebrid.LoadFromConfigFile(*path2Cfg)
	if err != nil {
		logging.LogFmt(logging.LOG_ERROR, "[main] unable to open or read app config file from [%s]: %v", *path2Cfg, err)
		logging.LogFmt(logging.LOG_WARN, "[main] using default settings")
		setts, _ = crebrid.LoadFromByteArr([]byte{})
	}
	me := crebrid.NewMainExecute(*setts)
	if !me.Init() {
		logging.LogFmt(logging.LOG_FATAL, "[main] failed to init main execute: %d", me.Status())
		os.Exit(2)
	}
	me.Start()
	// recognize signals
	go func() {
		s := <-sigs
		logging.LogFmt(logging.LOG_MAIN, "[main] receive os signal: %s", s)
		if (s == syscall.SIGQUIT) || (s == syscall.SIGTERM || s == syscall.SIGINT) {
			AppCleanup(me)
			os.Exit(0)
		} else {
			os.Exit(3)
		}
	}()
	// infinite main print loop
	for {
		time.Sleep(time.Millisecond * 50)
		if me.Status() != crebrid.SES_RUNNING {
			logging.LogFmt(logging.LOG_ERROR, "[main] main execute unexpectly stopped [%d] --> try restart", me.Status())
			if me.Status() == crebrid.SES_ERROR {
				me.Stop()
			}
			me = crebrid.NewMainExecute(*setts)
			if !me.Init() {
				logging.LogFmt(logging.LOG_FATAL, "[main] failed to init main execute: %d", me.Status())
				os.Exit(2)
			}
			me.Start()
		}
	}
}

func AppCleanup(me crebrid.Service) {
	logging.Log(logging.LOG_INFO, "[main] app cleanup called")
	me.Stop()
	for {
		if me.Status() == crebrid.SES_STOPPED {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
}
