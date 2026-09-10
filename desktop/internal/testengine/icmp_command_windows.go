package testengine

import (
	"os/exec"
	"syscall"
)

func configurePingCommand(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
