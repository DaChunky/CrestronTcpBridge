package crebrid

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/dachunky/crestrontcpbridge/pkg/ipc"
	"github.com/dachunky/crestrontcpbridge/pkg/logging"
)

const (
	CEC_OK    = iota
	CEC_FATAL = iota
)

type ServiceStatus int

const (
	SES_STOPPED ServiceStatus = iota
	SES_RUNNING
	SES_ERROR
)

type Service interface {
	Init() bool
	Start()
	Stop()
	Status() ServiceStatus
}

type mainExecute struct {
	ccc    CrestronControllerClient
	status ServiceStatus
	setts  CrebridDSettings
	doStop chan bool
	wait   sync.WaitGroup
}

func NewMainExecute(setts CrebridDSettings) Service {
	me := new(mainExecute)
	me.status = SES_STOPPED
	me.setts = setts
	return me
}

func (me *mainExecute) Init() bool {
	ccc, err := NewCrestronControllerClient(me.setts.IP, me.setts.Port)
	if err != nil {
		logging.LogFmt(logging.LOG_FATAL, "[service] creation of crestron controller client failed: %s", err)
		me.status = SES_ERROR
		return false
	}
	logging.LogFmt(logging.LOG_MAIN, "[service] successfully connected to controller: %s@%d", me.setts.IP, me.setts.Port)
	me.ccc = ccc
	me.ccc.SetAccessCode(me.setts.AccessCode)
	logging.Log(logging.LOG_DEBUG, "[service] try to send a get command")
	_, err = me.ccc.ToggleSwitch(0)
	if err != nil {
		logging.LogFmt(logging.LOG_FATAL, "[service] test command to server failed: %s", err)
		me.status = SES_ERROR
		return false
	}
	return true
}

func (me *mainExecute) Start() {
	me.doStop = make(chan bool)
	me.wait.Add(1)
	go me.execute()
}

func (me *mainExecute) Stop() {
	logging.Log(logging.LOG_MAIN, "[service] stopping main routine")
	me.doStop <- true
	logging.Log(logging.LOG_DEBUG, "[service] close stop channel")
	runtime.Gosched()
	me.wait.Wait()
	logging.Log(logging.LOG_DEBUG, "[service] wait excaped set status and return")
	me.status = SES_STOPPED
}

func (me *mainExecute) Status() ServiceStatus {
	return me.status
}

func (me *mainExecute) handleRequest(cc *ipc.ClientCommand) (*ipc.ServerResponse, error) {
	logging.LogFmt(logging.LOG_DEBUG, "[cmd handler] handle new request: %v", cc)
	sr := ipc.NewServerResponse()
	logging.Log(logging.LOG_DEBUG, "[cmd handler] create new response")
	sr.Cmd = cc.Cmd
	logging.Log(logging.LOG_DEBUG, "[cmd handler] setting command")
	var err error = nil
	ret := false
	switch sr.Cmd {
	case ipc.IC_REGISTER:
		sr.ID = cc.ID
	case ipc.IC_SINGLE, ipc.IC_MULTIPLE, ipc.IC_GET:
		containsStatusReq := sr.Cmd == ipc.IC_GET
		if sr.Cmd == ipc.IC_GET {
			ret, err = me.ccc.ToggleSwitch(0)
			if err != nil {
				logging.LogFmt(logging.LOG_ERROR, "toggle switch [%d] failed: %s", 0, err)
				if !ret {
					err = me.ccc.ReDial()
				} else {
					break
				}
			}
		} else {
			for _, sid := range cc.DigitalPorts {
				logging.LogFmt(logging.LOG_DEBUG, "[cmd handler] toggle switch %d", sid)
				isOn, err := me.ccc.ToggleSwitch(sid)
				logging.Log(logging.LOG_DEBUG, "[cmd handler] switch toggled")
				if err != nil {
					logging.LogFmt(logging.LOG_ERROR, "toggle switch [%d] failed: %s", sid, err)
					break
				}
				if sid > 0 {
					sr.DigitalPortInfo[sid] = isOn
				} else if sid == 0 {
					containsStatusReq = true
				}
			}
		}
		if containsStatusReq {
			for i, v := range me.ccc.GetSystemStatus().D {
				sr.DigitalPortInfo[i] = v > 0
			}
		}
	}
	if err != nil {
		logging.LogFmt(logging.LOG_DEBUG, "[cmd handler] request could not be handled: %v", err)
		return nil, err
	}
	logging.LogFmt(logging.LOG_DEBUG, "[cmd handler] request successfully handled: %v", cc)
	return sr, nil
}

func (me *mainExecute) execute() {
	me.status = SES_RUNNING
	defer func() {
		// me.wait.Done()
		logging.Log(logging.LOG_DEBUG, "[execute] wait group done")
	}()
	is := ipc.NewIpcServer(me.setts.IPCPort)
	go is.StartListening(me.handleRequest)
	logging.LogFmt(logging.LOG_MAIN, "[service] start to listen for IPC commands on port: %d", me.setts.IPCPort)
	defer is.Close()
	errTxt := ""
	aliveMsgTick := time.Now().Unix()
	for {
		switch {
		case <-me.doStop:
			logging.Log(logging.LOG_MAIN, "[service] main loop escaped")
			me.wait.Done()
			return
		default:
			if is.HasError() != nil {
				newErrTxt := fmt.Sprintf("[service] IPC listening interface results in an error --> escape main loop: %v", is.HasError())
				if newErrTxt != errTxt {
					errTxt = newErrTxt
					logging.Log(logging.LOG_FATAL, errTxt)
				}
				me.status = SES_ERROR
				return
			}
			time.Sleep(50 * time.Millisecond)
			if time.Now().Unix()-aliveMsgTick > 10 {
				aliveMsgTick = time.Now().Unix()
				logging.Log(logging.LOG_DEBUG, "[service] main thread still alive")
			}
		}
	}
}
