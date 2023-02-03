package crebrid

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/dachunky/crestrontcpbridge/pkg/ipc"
	"github.com/dachunky/crestrontcpbridge/pkg/logging"
)

type SystemStatus struct {
	D []int     `json:"d"`
	A []float64 `json:"a"`
}

func SystemStatusFromJSON(str string) (*SystemStatus, error) {
	ret := new(SystemStatus)
	err := json.Unmarshal([]byte(str), ret)
	if err != nil {
		return nil, err
	}
	return ret, err
}

const (
	system_state_toggle = 0
)

type CrestronControllerClient interface {
	// SetAccessCode for the controller
	SetAccessCode(accessCode string)
	// GetSystemStatus of the devices from the controller
	GetSystemStatus() *SystemStatus
	// ToggleSwitch with ID
	ToggleSwitch(switchID int) (bool, error)
	// Close the connection to the server
	Close()
	// Re-Dial close the current connection and re-dial
	ReDial() error
}

type crestronClient struct {
	ip         string
	port       int
	conn       net.Conn
	accessCode string
	curStatus  *SystemStatus
}

func NewCrestronControllerClient(ip string, port int) (CrestronControllerClient, error) {
	ccc := new(crestronClient)
	ccc.ip = ip
	ccc.port = port
	connStr := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.Dial("tcp", connStr)
	if err != nil {
		return nil, err
	}
	logging.LogFmt(logging.LOG_DEBUG, "[controller client] successfully connected to: %s", connStr)
	ccc.conn = conn
	ccc.curStatus = nil
	return ccc, nil
}

func (ccc *crestronClient) ReDial() error {
	logging.LogFmt(logging.LOG_INFO, "[controller client] close current connection on [%s:%d] and re-dial", ccc.ip, ccc.port)
	ccc.Close()
	connStr := fmt.Sprintf("%s:%d", ccc.ip, ccc.port)
	conn, err := net.Dial("tcp", connStr)
	if err != nil {
		return err
	}
	ccc.conn = conn
	logging.LogFmt(logging.LOG_DEBUG, "[controller client] successfully re-dial to: %s", connStr)
	return nil
}

func (ccc *crestronClient) Close() {
	logging.Log(logging.LOG_INFO, "[controller client] close connection to server")
	ccc.conn.Close()
}

func (ccc *crestronClient) SetAccessCode(accessCode string) {
	ccc.accessCode = accessCode
}

func (ccc *crestronClient) GetSystemStatus() *SystemStatus {
	return ccc.curStatus
}

func (ccc *crestronClient) UpdateSystemStatus() error {
	_, err := ccc.ToggleSwitch(system_state_toggle)
	return err
}

func (ccc *crestronClient) waitForControllerResponse(timeout int) ([]byte, error) {
	var resp []byte
	waitChan := make(chan bool)
	err := fmt.Errorf("timeout received within [%ds]", timeout)
	logging.LogFmt(logging.LOG_DEBUG, "[WAIT] wait for response for %dms", timeout)
	defer close(waitChan)
	go func() {
		resp, err = ipc.ReadUntilEOF(bufio.NewReader(ccc.conn))
		waitChan <- true
	}()
	responseReceived := false
	startWait := time.Now()
	for {
		waitSince := 0
		switch {
		case <-waitChan:
			responseReceived = true
		default:
			waitSince = int(time.Since(startWait).Milliseconds())
			time.Sleep(time.Millisecond * 250)
			logging.LogFmt(logging.LOG_DEBUG, "[WAIT] wait since %dms", waitSince)
		}
		if responseReceived || (waitSince >= timeout) {
			break
		}
	}
	if responseReceived {
		return resp, nil
	} else {
		return nil, err
	}
}

func (ccc *crestronClient) ToggleSwitch(switchID int) (bool, error) {
	cmdStr := fmt.Sprintf("%s%3.3d", ccc.accessCode, switchID)
	logging.LogFmt(logging.LOG_DEBUG, "[controller client] sending command: %s", cmdStr)
	_, err := ccc.conn.Write([]byte(cmdStr))
	if err != nil {
		logging.LogFmt(logging.LOG_ERROR, "[controller client] failed to write on connection: %s", ccc.conn.RemoteAddr().String())
		return false, err
	}
	logging.Log(logging.LOG_DEBUG, "[controller client] waiting for response")
	// create a timeout for the response. normally it has to be responded by the controller within ms
	// channel to wait for the response
	resp, err := ccc.waitForControllerResponse(1000)
	if err != nil {
		if ccc.curStatus == nil {
			return false, err
		} else {
			return true, err
		}
	}
	logging.LogFmt(logging.LOG_DEBUG, "[controller client] receive response: %s", string(resp))
	ss, err := SystemStatusFromJSON(string(resp))
	if err != nil {
		return false, err
	}
	logging.LogFmt(logging.LOG_DEBUG, "[controller client] current system status: %v", ss)
	ccc.curStatus = ss
	ret := false
	if switchID < 1 {
		ret = true
	} else if ok := switchID <= len(ss.D); ok {
		ret = ss.D[switchID-1] > 0
		logging.LogFmt(logging.LOG_DEBUG, "[controller client] switch ID [%d] is set to: %v", switchID, ret)
	}
	return ret, nil
}
