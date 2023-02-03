package crebrid

import (
	"os"

	"github.com/dachunky/crestrontcpbridge/pkg/logging"
	"gopkg.in/ini.v1"
)

// {"serverIP":"192.168.178.32","port":43123,"accessCode":"3H34GJ67NH"}

type CrebridDSettings struct {
	IP         string
	Port       int
	IPCPort    int
	AccessCode string
}

type configFileKey int

const (
	cfk_ip configFileKey = iota
	cfk_port
	cfk_ipc_port
	cfk_access_code
)

var configFileKeyString = map[configFileKey]string{
	cfk_ip:          "ip",
	cfk_port:        "port",
	cfk_ipc_port:    "ipcPort",
	cfk_access_code: "accessCode",
}

func LoadFromByteArr(data []byte) (*CrebridDSettings, error) {
	iniFl, err := ini.Load(data)
	if err != nil {
		return nil, err
	}
	cs := new(CrebridDSettings)
	sec := iniFl.Section("")
	for enm, key := range configFileKeyString {
		logging.LogFmt(logging.LOG_DEBUG, "found entry: %s", key)
		switch enm {
		case cfk_ip:
			cs.IP = sec.Key(key).MustString("192.168.178.32")
		case cfk_port:
			cs.Port = sec.Key(key).MustInt(43123)
		case cfk_ipc_port:
			cs.IPCPort = sec.Key(key).MustInt(65432)
		case cfk_access_code:
			cs.AccessCode = sec.Key(key).MustString("3H34GJ67NH")
		}
	}
	return cs, nil
}

func LoadFromConfigFile(path2File string) (*CrebridDSettings, error) {
	data, err := os.ReadFile(path2File)
	if err != nil {
		// "read" default settings, if file is not available
		logging.LogFmt(logging.LOG_WARN, "[SETTINGS] unable to read settings from: %s", path2File)
		data = []byte{}
	}
	cs, err := LoadFromByteArr(data)
	if err != nil {
		return nil, err
	}
	return cs, nil
}
