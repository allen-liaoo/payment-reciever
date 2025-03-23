package main

import (
	"os/exec"
	"syscall"
	"time"
)

// Chain's RPC server always live at 127.0.0.1:8545
func StartChain() (*exec.Cmd, error) {
	cmd := exec.Command("npx", "hardhat", "node")
	err := cmd.Start()
	// has to wait a bit since hardhat spawns a child process that runs the rpc server
	// TODO: Better way to do this?
	time.Sleep(3 * time.Second)
	return cmd, err
}

func StopChain(chain_cmd *exec.Cmd) error {
	return syscall.Kill(chain_cmd.Process.Pid, syscall.SIGTERM)
}
