package emulator

import (
	"log"
	"os/exec"
)

func printInvocation(cmd *exec.Cmd) {
	if PrintInvocations {
		log.Println(cmd.String())
	}
}
