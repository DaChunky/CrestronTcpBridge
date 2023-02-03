package crebri

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dachunky/crestrontcpbridge/pkg/crebrid"
	"github.com/dachunky/crestrontcpbridge/pkg/ipc"
	"github.com/dachunky/crestrontcpbridge/pkg/logging"
)

func interactive(ic ipc.IpcClient) {
	fmt.Println("-------- start crebri client ---------")
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("client:>")
		text, _ := reader.ReadString('\n')
		if strings.TrimSpace(string(text)) == "q" {
			logging.Log(logging.LOG_DEBUG, "TCP client exiting...")
			break
		}
		port, err := strconv.Atoi(text[:(len(text) - 1)])
		if err != nil {
			fmt.Printf("client:> invalid input [%s]. Only number are allowed", text)
			continue
		}
		cc := ipc.NewClientCommand()
		cc.ID = ic.ClientID()
		cc.Cmd = ipc.IC_SINGLE
		cc.AddDigitalPorts(port)
		resp, err := ic.SendCommand(cc)
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Printf("client:> %v\n", resp.DigitalPortInfo[port])
		time.Sleep(250)
	}
}

func Execute() error {
	logging.Log(logging.LOG_MAIN, "[execute] start client")
	// read command line arguments
	cmdArgs, err := ParseAppArguments(os.Args[1:])
	if err != nil {
		return err
	}
	logging.LogFmt(logging.LOG_MAIN, "[execute] arguments parsed: %v", cmdArgs)
	// read settings from /etc/crebrid/crebrid.conf
	setts, err := crebrid.LoadFromConfigFile("/etc/crebrid/crebrid.conf")
	if err != nil {
		return err
	}
	logging.LogFmt(logging.LOG_MAIN, "try to connect to service: %s:%d", setts.IP, setts.IPCPort)
	// connect to service via ipc
	ic, err := ipc.RegisterClient(cmdArgs.ServiceIP, setts.IPCPort)
	if err != nil {
		return err
	}
	defer ic.CloseConnection()
	switch cmdArgs.Cmd {
	case CCT_SET:
		cc := ipc.NewClientCommand()
		cc.ID = ic.ClientID()
		cc.Cmd = ipc.IC_SINGLE
		cc.AddDigitalPorts(cmdArgs.Port)
		resp, err := ic.SendCommand(cc)
		if err != nil {
			return err
		}
		if cmdArgs.Port > 0 {
			if resp.DigitalPortInfo[cmdArgs.Port] {
				fmt.Println("ON")
			} else {
				fmt.Println("OFF")
			}
		} else {
			s := resp.TransformSystemState()
			fmt.Println(s)
		}
		return nil
	case CCT_GET:
		cc := ipc.NewClientCommand()
		cc.ID = ic.ClientID()
		cc.Cmd = ipc.IC_GET
		cc.AddDigitalPorts(cmdArgs.Port)
		resp, err := ic.SendCommand(cc)
		if err != nil {
			return err
		}
		if cmdArgs.Port > 0 {
			if resp.DigitalPortInfo[cmdArgs.Port] {
				fmt.Println("ON")
			} else {
				fmt.Println("OFF")
			}
		} else {
			s := resp.TransformSystemState()
			fmt.Println(s)
		}
		return nil
	}
	interactive(ic)
	return nil
}
