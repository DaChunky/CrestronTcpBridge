package ipc

import (
	"bufio"
	"fmt"
	"net"

	"github.com/dachunky/crestrontcpbridge/pkg/logging"
)

// IpcClient provides methods to communiacate with a service
type IpcClient interface {
	// ClientID
	ClientID() string
	// SendCommand to a service
	SendCommand(cc *ClientCommand) (*ServerResponse, error)
	// CloseConnection to the service
	CloseConnection() error
}

// ipcClient represents a registered IPC client
type ipcClient struct {
	id      string
	ipcConn net.Conn
}

// RegisterClient
func RegisterClient(ip string, port int) (IpcClient, error) {
	connStr := fmt.Sprintf("%s:%d", ip, port)
	// connect to the service
	cs, err := net.Dial("tcp", connStr)
	if err != nil {
		return nil, err
	}
	logging.LogFmt(logging.LOG_INFO, "[IPCCLIENT]: successfully connected to %s", connStr)
	// receive ID from the server
	cc := NewClientCommand()
	cc.Cmd = IC_REGISTER
	cmdData, err := cc.GetCommand2Send()
	if err != nil {
		return nil, err
	}
	_, err = cs.Write(cmdData)
	if err != nil {
		logging.LogFmt(logging.LOG_ERROR, "[IPCCLIENT] failed to write in connection stream: %v", err)
		return nil, err
	}
	response, err := ReadUntilEOF(bufio.NewReader(cs)) //.ReadBytes('\n')
	if err != nil {
		logging.LogFmt(logging.LOG_ERROR, "[IPCCLIENT] receive error register response: %v", err)
		return nil, err
	}
	logging.LogFmt(logging.LOG_INFO, "[IPCCLIENT]: successfully registered @ %s", connStr)
	sr, err := ServerResponseFromResponse(response)
	if err != nil {
		return nil, err
	}
	if sr.Cmd != IC_REGISTER {
		return nil, fmt.Errorf("[IPCCLIENT]: receive unexpected register command from server: %d", sr.Cmd)
	}
	ret := new(ipcClient)
	ret.id = sr.ID
	ret.ipcConn = cs
	return ret, nil
}

func (ic *ipcClient) ClientID() string {
	return ic.id
}

func (ic *ipcClient) SendCommand(cc *ClientCommand) (*ServerResponse, error) {
	if ic.ipcConn == nil {
		return nil, fmt.Errorf("cannot send request with an empty server connection")
	}
	cc.ID = ic.id
	data, err := cc.GetCommand2Send()
	if err != nil {
		return nil, err
	}
	logging.LogFmt(logging.LOG_DEBUG, "[IPCCLIENT] send data to IPC server: %v", data)
	_, err = ic.ipcConn.Write(data)
	if err != nil {
		logging.LogFmt(logging.LOG_ERROR, "[IPCCLIENT] failed to write in connection stream: %v", err)
		return nil, err
	}
	logging.Log(logging.LOG_DEBUG, "[IPCCLIENT] data sent --> waiting for response")
	resp, err := ReadUntilEOF(bufio.NewReader(ic.ipcConn))
	if err != nil {
		logging.LogFmt(logging.LOG_ERROR, "[IPCCLIENT] receive error response: %v", err)
		return nil, err
	}
	logging.LogFmt(logging.LOG_DEBUG, "[IPCCLIENT] response received: %v", resp)
	sr, err := ServerResponseFromResponse(resp)
	if err != nil {
		return nil, err
	}
	if cc.Cmd != sr.Cmd {
		return nil, fmt.Errorf("receive = %d but want %d", sr.Cmd, cc.Cmd)
	}
	logging.LogFmt(logging.LOG_DEBUG, "[IPCCLIENT] response successfully verified --> returning: %v", sr)
	return sr, nil
}

func (ic *ipcClient) CloseConnection() error {
	logging.Log(logging.LOG_INFO, "[IPCCLIENT] unregister from server")
	fmt.Fprint(ic.ipcConn, "q")
	logging.Log(logging.LOG_INFO, "[IPCCLIENT] sent quit command 'q'")
	err := ic.ipcConn.Close()
	return err
}
