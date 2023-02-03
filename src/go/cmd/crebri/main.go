package main

import (
	"fmt"
	"os"

	"github.com/dachunky/crestrontcpbridge/pkg/crebri"
	"github.com/dachunky/crestrontcpbridge/pkg/logging"
)

func main() {
	logging.Log(logging.LOG_MAIN, "[main] application started")
	err := crebri.Execute()
	if err != nil {
		fmt.Printf("client execute failed: %v\n", err)
		logging.LogFmt(logging.LOG_FATAL, "[main] main execute failed: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}
