package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/ghoshRitesh12/montre/lib"
)

func main() {
	montre := lib.Init()
	montre.StartWatching()

	os.Exit(0)

	fmt.Println("(go pid) or ppid", os.Getpid())

	cmd := exec.Command("node", "index.js")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Println("some error occured:", err.Error())
		return
	}

	fmt.Println("node pid", cmd.Process.Pid)

	if err := cmd.Wait(); err != nil {
		fmt.Println("process finished with error:", err.Error())
		return
	}

	fmt.Println(cmd.Path)
}
