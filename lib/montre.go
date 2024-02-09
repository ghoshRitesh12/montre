package lib

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/radovskyb/watcher"
)

type Montre struct {
	Args     []string
	Filename string
	Pwd      string
	Child    *exec.Cmd
	watcher  *watcher.Watcher
	// to watch files(default .go) that match this pattern
	AllowList string
	// to ignore files that match this pattern
	IgnoreList string
	// true by default
	IgnoreHiddenFiles bool
}

func Init() *Montre {
	montre := &Montre{
		Args:              os.Args,
		Filename:          os.Args[1],
		AllowList:         "/**/*.go",
		IgnoreHiddenFiles: true,
		watcher:           watcher.New(),
	}

	if configFileExist() {
		populateConfig(montre)
	}

	_, err := os.Stat(montre.Filename)
	if errors.Is(err, os.ErrNotExist) {
		panic(ErrMainFileNotFound)
	}

	pwd, err := os.Getwd()
	if err != nil {
		panic("error while setting present working directory")
	}
	montre.Pwd = pwd

	fmt.Println(YellowColor+"[montre] watching path(s):", montre.AllowList+ResetColor)
	// fmt.Println(YellowColor + "[montre] watching extension(s): go" + ResetColor)
	fmt.Println(GreenColor + "[montre] starting `go run main.go`" + ResetColor)

	return montre
}

func (m *Montre) StartWatching() {
	m.watcher.SetMaxEvents(1)
	m.watcher.FilterOps(
		watcher.Write,
		watcher.Rename,
		watcher.Remove,
		watcher.Move,
	)
	m.watcher.IgnoreHiddenFiles(m.IgnoreHiddenFiles)

	fmt.Printf("%+v\n", m)

	if err := m.watcher.Ignore([]string{m.IgnoreList}...); err != nil {
		panic("error while watching recursive path")
	}
	if err := m.watcher.AddRecursive(m.AllowList); err != nil {
		panic("error while watching recursive path")
	}

	go m.ListeningEvents()

	m.Reload()
}

func (m *Montre) Reload() {
	if m.Child != nil {
		err := m.Child.Process.Kill()
		if err != nil {
			panic(ErrRestartingProcess)
		}
	}

	cmd := exec.Command("node", m.Filename)
	// cmd.StderrPipe()
	// cmd.StdoutPipe()
	// cmd.Stdin = os.Stdin
	// cmd.Stderr = os.Stderr
	// cmd.Stdout = os.Stdout

	m.Child = cmd

	if err := m.Child.Start(); err != nil {
		panic("some error occured: " + err.Error())
	}

	if err := m.Child.Wait(); err != nil {
		fmt.Println("process finished with error:", err.Error())
		return
	}
}

func (m *Montre) ListeningEvents() {
	for {
		select {
		case event := <-m.watcher.Event:
			fmt.Println("event", event)
		case err := <-m.watcher.Error:
			fmt.Println(err)
		case <-m.watcher.Closed:
			return
		}
	}
}
