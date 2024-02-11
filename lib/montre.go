package lib

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type Montre struct {
	args    []string // could be expanded upon in future
	pwd     string
	config  Config
	child   *exec.Cmd
	watcher *fsnotify.Watcher
	blocker chan struct{}
}

func Init() *Montre {
	montre := &Montre{
		args: os.Args,
		config: Config{
			MainFile:  os.Args[1],
			WatchExts: []string{".go"},
		},
	}

	if configFileExist() {
		populateConfig(montre)
	}

	mainFileInfo, err := os.Stat(montre.config.MainFile)
	if errors.Is(err, os.ErrNotExist) {
		log.Fatalln(ErrMainFileNotFound)
	}
	if mainFileInfo.IsDir() {
		log.Fatalln(ErrMainFileIsDir)
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalln(ErrSettingPWD)
	}
	montre.pwd = pwd

	return montre
}

func (m *Montre) StartWatching() {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln(ErrWatcherSetup)
	}
	m.watcher = w

	defer m.watcher.Close()

	go m.initListenEvents()
	go m.acceptCommands()

	walkErr := filepath.WalkDir(m.pwd, func(path string, file fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if file.IsDir() && slices.Contains(m.config.IgnoreDirs, file.Name()) {
			return filepath.SkipDir
		}

		if !slices.Contains(m.config.WatchExts, filepath.Ext(path)) {
			return nil
		}

		addErr := m.watcher.Add(path)
		if addErr != nil {
			return nil
		}

		return nil
	})

	if walkErr != nil {
		log.Fatalln(ErrWalkingFS)
	}

	fmt.Println(MONTRE_LOG + YellowLog("watching extension(s): "+strings.Join(m.config.WatchExts, ",")))
	fmt.Println(MONTRE_LOG + YellowLog("ignoring folder(s): "+strings.Join(m.config.IgnoreDirs, ", ")))
	fmt.Println(MONTRE_LOG + YellowLog("to restart watcher enter ") + GreenLog("`rs`"))
	fmt.Println(MONTRE_LOG + YellowLog("to quit watching enter ") + RedLog("`q`"))
	fmt.Println(MONTRE_LOG + GreenLog("starting `go run main.go`"))

	m.reload()
	<-m.blocker
}

func (m *Montre) reload() {
	m.quitChildProcess(ErrRestartChildProcess)

	cmd := exec.Command("go", "run", m.config.MainFile)
	m.child = cmd

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := m.child.Start(); err != nil {
		log.Fatalln(ErrStartChildProcess)
	}

	fmt.Println(MONTRE_LOG + GreenLog("waiting for further changes"))
}

func (m *Montre) initListenEvents() {
	for {
		select {
		case err, ok := <-m.watcher.Errors:
			if !ok {
				log.Fatalln(RedLog(err.Error()))
				return
			}
		case _, ok := <-m.watcher.Events:
			// _  is event struct
			if !ok {
				return
			}
			m.reload()
			// fmt.Println(event.Op, event.Name)
		}
	}
}

func (m *Montre) acceptCommands() {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Err() != nil {
		log.Fatalln(ErrWatcherSetup)
	}

	for {
		if scanner.Scan() {
			cmd := scanner.Text()
			switch cmd {
			case "rs":
				m.reload()
			case "q":
				m.quitChildProcess(ErrKillChildProcess)
				os.Exit(0)
			}
		}
	}
}

func (m *Montre) quitChildProcess(errStr error) {
	if m.child != nil {
		err := m.child.Process.Kill()
		if err != nil {
			log.Fatalln(RedLog(errStr.Error() + " " + err.Error()))
		}
	}
}
