package crebri

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/dachunky/crestrontcpbridge/pkg/logging"
)

type CommandType int

const (
	CCT_SERVER CommandType = iota
	CCT_SET
	CCT_GET
	CCT_INTERACTIVE
)

var commandTypeStr = map[CommandType]string{
	CCT_SERVER: "server",
	CCT_SET:    "set",
	CCT_GET:    "get",
}

type RegisterType int

const (
	CRT_DIGITAL RegisterType = iota
	CRT_ANALOG
	CRT_STRING
)

var registerTypeStr = map[RegisterType]string{
	CRT_DIGITAL: "d",
	CRT_ANALOG:  "a",
	CRT_STRING:  "s",
}

var registerStrToType = map[string]RegisterType{
	"d": CRT_DIGITAL,
	"a": CRT_ANALOG,
	"s": CRT_STRING,
}

type ParsedArguments struct {
	ServiceIP string
	Cmd       CommandType
	Register  RegisterType
	Port      int
	ValueStr  string
	ValueInt  int
}

func (pa *ParsedArguments) asStringLine() string {
	return fmt.Sprintf("IP:%s->Cmd:%s->Reg:%s->Port:%d->Str:%s->Int:%d", pa.ServiceIP, commandTypeStr[pa.Cmd], registerTypeStr[pa.Register], pa.Port, pa.ValueStr, pa.ValueInt)
}

func ParseAppArguments(args []string) (*ParsedArguments, error) {
	ret := new(ParsedArguments)
	ret.ServiceIP = "localhost"
	ret.Cmd = CCT_INTERACTIVE
	ret.Register = CRT_DIGITAL
	ret.Port = 0
	ret.ValueInt = 0
	ret.ValueStr = ""
	arrLen := len(args)
	if arrLen < 1 {
		logging.Log(logging.LOG_MAIN, "no command line arguments provided. starting interactive mode")
		return ret, nil
	}
	servFls := flag.NewFlagSet(commandTypeStr[CCT_SERVER], flag.ExitOnError)
	servIp := servFls.String("ip", "localhost", "define crebrid service ip. default is localhost")
	setFls := flag.NewFlagSet(commandTypeStr[CCT_SET], flag.ExitOnError)
	setRegType := setFls.String("reg", "d", "register type to set. default is digital")
	setPort := setFls.Int("port", -1, "port to set")
	setValue := setFls.String("value", "", "value to set in case of analog or serial command")
	getFls := flag.NewFlagSet(commandTypeStr[CCT_GET], flag.ExitOnError)
	getRegType := getFls.String("reg", "d", "register type to get. default is digital")
	getPort := getFls.Int("port", -1, "port to get")
	argIdx := 0
	correctedArgs := make([]string, len(args))
	for idx, arg := range args {
		logging.LogFmt(logging.LOG_DEBUG, "[%d] %s", idx, arg)
		if strings.Contains(arg, "port") {
			portSplit := strings.Split(arg, "=")
			portNr := portSplit[1]
			rmIdx := 0
			for _, c := range portSplit[1] {
				if c == '0' {
					rmIdx++
				} else {
					break
				}
			}
			if rmIdx > 0 {
                                if rmIdx < len(portSplit[1]) {
                                	portNr = portNr[rmIdx:]
                                }
			}
			correctedArgs[idx] = fmt.Sprintf("%s=%s", portSplit[0], portNr)
		} else {
			correctedArgs[idx] = arg
		}
	}
	if args[argIdx] == commandTypeStr[CCT_SERVER] {
		if arrLen < 2 {
			return nil, fmt.Errorf("detect 'server' argument but IP info is missing")
		}
		// receive server seetings
		servFls.Parse(correctedArgs[1:2])
		ret.ServiceIP = *servIp
		logging.LogFmt(logging.LOG_INFO, "set service IP to: %s", ret.ServiceIP)
		argIdx = argIdx + 2
		if arrLen <= argIdx {
                        logging.Log(logging.LOG_DEBUG, "no more arguments provided")
			return ret, nil
		}
	}
        logging.LogFmt(logging.LOG_DEBUG, "perform cmd [%s]", correctedArgs[argIdx])
	switch correctedArgs[argIdx] {
	case commandTypeStr[CCT_SET]:
		ret.Cmd = CCT_SET
		setFls.Parse(correctedArgs[(argIdx + 1):])
		reg, ok := registerStrToType[*setRegType]
		if !ok {
			return nil, fmt.Errorf("unknown register type: %s", *setRegType)
		}
		ret.Register = reg
		if *setPort < 0 {
			return nil, fmt.Errorf("invalid port: %d", *setPort)
		}
		ret.Port = *setPort
		if *setValue != "" {
			switch ret.Register {
			case CRT_ANALOG:
				valInt, err := strconv.Atoi(*setValue)
				if err != nil {
					return nil, fmt.Errorf("invalid analog value provided [%s]: %v", *setValue, err)
				}
				ret.ValueInt = valInt
			case CRT_STRING:
				ret.ValueStr = *setValue
			}
		}
	case commandTypeStr[CCT_GET]:
		ret.Cmd = CCT_GET
		getFls.Parse(correctedArgs[(argIdx + 1):])
		reg, ok := registerStrToType[*getRegType]
		if !ok {
			return nil, fmt.Errorf("unknown register type: %s", *setRegType)
		}
		ret.Register = reg
		if *getPort < 0 {
			return nil, fmt.Errorf("invalid port: %d", *setPort)
		}
		ret.Port = *getPort
	default:
		return nil, fmt.Errorf("either provide no arguments for interactive mode or set or get")
	}
	logging.LogFmt(logging.LOG_DEBUG, "return command: %s", ret.asStringLine())
	return ret, nil
}
