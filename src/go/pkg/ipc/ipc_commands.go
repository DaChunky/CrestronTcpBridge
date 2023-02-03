package ipc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/dachunky/crestrontcpbridge/pkg/logging"
	"github.com/google/uuid"
)

const (
	// IC_REGISTER regist a client by the service and expects an ID in return
	IC_REGISTER = iota
	// IC_SINGLE send a command for one single item
	IC_SINGLE
	// IC_MULTIPLE send a command for multiple items
	IC_MULTIPLE
	// IC_GET the state of an item
	IC_GET
)

const (
	CLIENT_QUIT_COMMAND = "q"
)

// ClientCommand holds needed information about a client request
type ClientCommand struct {
	Cmd          int    `json:"cmd"`
	ID           string `json:"id"`
	DigitalPorts []int  `json:"digitalPorts"`
}

func (cc *ClientCommand) AddDigitalPorts(ports ...int) {
	cc.DigitalPorts = append(cc.DigitalPorts, ports...)
}

func (cc *ClientCommand) serialize() ([]byte, error) {
	var err error
	defer catchError(err)
	res, err := json.Marshal(cc)
	logging.LogFmt(logging.LOG_DEBUG, "data to encrypt: %s", string(res))
	if err != nil {
		res = nil
	} else {
		res = encrypt(res, aesPassphrase)
	}
	return res, err
}

func (cc *ClientCommand) deserialize(data []byte) error {
	var err error
	defer catchError(err)
	decData := decrypt(data, aesPassphrase)
	logging.LogFmt(logging.LOG_DEBUG, "[DESERIALIZE] received decrypted and deserialized data: %s", string(decData))
	err = json.Unmarshal(decData, cc)
	return err
}

// GetCommand2Send creates an encypted IPC command
func (cc *ClientCommand) GetCommand2Send() ([]byte, error) {
	data, err := cc.serialize()
	if err != nil {
		return nil, err
	}
	return data, nil
}

// NewClientCommand with initialized digital port slice
func NewClientCommand() *ClientCommand {
	res := new(ClientCommand)
	res.DigitalPorts = make([]int, 0)
	return res
}

// ClientCommandFromRequest by encrypted data. Returns an error if deserialization failed
func ClientCommandFromRequest(data []byte) (*ClientCommand, error) {
	cc := NewClientCommand()
	err := cc.deserialize(data)
	if err != nil {
		return nil, err
	}
	return cc, nil
}

// ServerResponse from an IPC request
type ServerResponse struct {
	Cmd             int          `json:"cmd"`
	ID              string       `json:"id"`
	DigitalPortInfo map[int]bool `json:"digitalPortInfo"`
	ResponseID      string       `json:"responseId"`
}

func (sr *ServerResponse) serialize() ([]byte, error) {
	var err error
	defer catchError(err)
	ret, err := json.Marshal(sr)
	if err != nil {
		ret = nil
	} else {
		ret = encrypt(ret, aesPassphrase)
	}
	return ret, err
}

func (sr *ServerResponse) deserialize(data []byte) error {
	var err error
	defer catchError(err)
	decData := decrypt(data, aesPassphrase)
	logging.LogFmt(logging.LOG_DEBUG, "[DESERIALIZE] received decrypted and deserialized data: %s", string(decData))
	err = json.Unmarshal(decData, sr)
	return err
}

// GetResponse2Send as encrypted and serialized data stream
func (sr *ServerResponse) GetResponse2Send() ([]byte, error) {
	data, err := sr.serialize()
	if err != nil {
		return nil, err
	}
	return data, nil
}

// TransformSystemState shows the status of all switches of the system
func (sr *ServerResponse) TransformSystemState() string {
	stateArr := make([]string, len(sr.DigitalPortInfo))
	for i, b := range sr.DigitalPortInfo {
		if b {
			stateArr[i] = "ON"
		} else {
			stateArr[i] = "OFF"
		}
	}
	return strings.Join(stateArr, ",")
}

// NewServerResponse with an initialized digital port info map
func NewServerResponse() *ServerResponse {
	res := new(ServerResponse)
	res.DigitalPortInfo = make(map[int]bool)
	res.ResponseID = uuid.NewString()
	return res
}

// ServerResponseFromResponse decrypted and deserialize a response from an IPC server
func ServerResponseFromResponse(data []byte) (*ServerResponse, error) {
	sr := NewServerResponse()
	err := sr.deserialize(data)
	if err != nil {
		return nil, err
	} else {
		return sr, nil
	}
}

// ReadUntilEOF
func ReadUntilEOF(reader *bufio.Reader) ([]byte, error) {
	ret := make([]byte, 0)
	block := 1024
	zeroLengthRetry := 0
	logging.Log(logging.LOG_DEBUG, "[ReadUntilEOF] try to read from buffer")
	for {
		buf := make([]byte, block)
		n, err := reader.Read(buf)
		logging.LogFmt(logging.LOG_DEBUG, "[ReadUntilEOF] read [%d] bytes", n)
		ret = append(ret, buf[:n]...)
		if err != nil {
			if err == io.EOF {
				if n < 1 {
					zeroLengthRetry = zeroLengthRetry + 1
					if zeroLengthRetry < 10 {
						time.Sleep(100 * time.Millisecond)
						continue
					}
					logging.LogFmt(logging.LOG_DEBUG, "[ReadUntilEOF] read EOF with zero content [%d] times", zeroLengthRetry)
					return nil, fmt.Errorf("got [%d] times EOF but no data", zeroLengthRetry)
				}
				logging.LogFmt(logging.LOG_DEBUG, "[ReadUntilEOF] read EOF [%d bytes]", n)
				return ret, nil
			}
			logging.LogFmt(logging.LOG_DEBUG, "[ReadUntilEOF] got error not equal to EOF: %v", err)
			return nil, err
		}
		logging.LogFmt(logging.LOG_DEBUG, "[ReadUntilEOF] read [%d] bytes from stream", n)
		if n < block {
			return ret, nil
		}
	}
}
