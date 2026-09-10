//go:build !windows

package testengine

import "os/exec"

func configurePingCommand(cmd *exec.Cmd) {}
