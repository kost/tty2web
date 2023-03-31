// httpsc in Go. Copyright (C) Kost. Distributed under MIT.

package server

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func ForkAndExecute(cc *SCConfig, ptype string, cmdstr string) {
	var cmd *exec.Cmd

	env := os.Environ()
	newEnv := []string{
		fmt.Sprintf("T2W_CMD=%s", ptype),
		fmt.Sprintf("T2W_SC=%s", cmdstr),
	}
	env = append(env, newEnv...)

	exfn, erre := os.Executable()
	if erre != nil {
		log.Fatal(erre)
		return
	}
	log.Printf("Executing %s with T2W_CMD=%s T2W_SC=%s", exfn, ptype, cmdstr)

	cmd = exec.Command(exfn, "--childsc")

	cmd.Env = env
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	return
}
