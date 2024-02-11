<p align="center">
  <img
    src="https://github.com/ghoshRitesh12/montre/assets/101876769/205953f8-e1f2-49ec-98d3-22e341eb883a"
    alt="montre_go"
  />
</p>

## How does nodemon work?

> I have always had this question because I wanted to know how nodemon detects changes made in a file system or a file tree and then restarts the whole node process.
>
> How does it get the entry point program's process id? Does it use the "node" command internally?

When Googling, I came across a [great article](https://www.pankajtanwar.in/blog/have-you-ever-thought-how-nodemon-works-internally-lets-build-our-own-nodemon-in-under-10-minutes) that portrays a good view of how nodemon works under the hood, or maybe how it's supposed to work.
Before reading the article, I had some prejudice and was wondering how it worked. I had a rough idea that obviously it's calling the "node" binary underneath, but how does it detect changes to the file system, and how does it reload or restart the server or node process based on an event?

Well, all of those questions and more were answered once I read the previously mentioned article, and obviously my mind was blown. I knew that it was calling the node process but didn't have any idea that it was actually forking a node `child_process`. So basically, `nodemon` acts as a wrapper around the `node` process, and on any relevant changes to the file system, it restarts the server or process by killing the existing child process and then `spawn`ing a new child process.

Nodemon watches on specific directories and files by default and also ignores some obvious ones like `node_modules`, its config can also be expanded upon by providing a `nodemon.json` file, which nodemon uses internally. It also uses the `package.json` for gathering the project info or other types of relevant information.

It uses the `chokidar` npm library for detecting file or directory changes because apparently the inbuilt `fs.watch` in node is faulty as it fires unnecessary events, which are sometimes not appropriate. I also found out that `chokidar` is itself a wrapper around `fs.watch`, exposing an optimized and appropriate API for detecting event-based file system changes.

## To set out on an adventure

> Sneaking in a little bit of One Piece reference

In order to explore and learn more, I set out on an adventure to make a nodemon equivalent tool for Go, so that I don't have to manually restart, say a server or a process when developing. I already know that there are many available tools that could make my life easier, but I was up for a challenge big time.

For me to make a nodemon like CLI tool for Go, I needed to figure out two things really:

- A **watcher** for file system events
- How to kill and start a **child process** over again

1. For the **watcher** I was hopping around [watcher](https://github.com/radovskyb/watcher) and [fsnotify](https://github.com/fsnotify/fsnotify). I saw that `watcher` already had a recursive watching capability, so I wanted to use that but I couldn't end up setting it up perhaps for the limited documentation for my use case. I ended up settling on fsnotify for its rich documentation. I also came across `inotify` Linux API, but I was too scared to deal with C :) (hopefully someday it won't be so).

2. I needed to know how I could execute a **child process**, coming from a node background I know it's as easy as using a `child_process.spawn` or `child_process.fork` for spinning up a new child process. Eventually, I read through the standard library documentation of Go and stumbled upon the `os/exec` package and I already knew that's what I wanted. I further Googled it up for more documentation and came across this [gobyexample page](https://gobyexample.com/execing-processes) and this [zetcode page](https://zetcode.com/golang/exec-command/) which provided me with solid insights on using this API.

## The Implementation

### Naming the nodemon clone

We all know that coding isn't that hard; it's naming things that takes the most time. Initially, my lazy self was thinking of naming it **gomon**, but was brought to enlightenment as I googled the name up and a package already existed with the same name, so I thought hard, and I was inspired by how [Evan You](https://evanyou.me/) the creator of Vue and Vite, on how he came up with the name Vite. Well, **Vite** stands for **quick** in French, and it struck me. I didn't waste a single minute and just converted the word **watch** to French and what I got was **montre** and I was like, it sounds okayish, so I rolled along with it.

### The Code

_Here is the [repo](https://github.com/ghoshRitesh12/montre) if this article is already too long :3_

Since nodemon used a `nodemon.json` file for extended config, I also thought that `montre` should have that too. Obviously, it will be called **`montre.json`** duh!

I wanted to have custom errors and wanted those errors and normal logs to have a bit of flair, so I pulled up a `bash` cheat sheet and found how to have colorful logs by prefixing certain sequences of characters (idk the term for these characters). \
I went with these custom errors:

```go
var (
  ErrMainFileNotFound    error = getErr("specified main file not found")
  ErrNoMainFile          error = getErr("no main file specified in config or command")
  ErrEmptyAllowList      error = getErr("cannot have empty allow list")
  ErrReadingConfigFile   error = getErr("error reading config file")
  ErrParsingConfigFile   error = getErr("error while parsing config file")
  ErrStartChildProcess   error = getErr("error while starting program")
  ErrRestartChildProcess error = getErr("error while restarting program")
  ErrKillChildProcess    error = getErr("error while killing program")
  ErrMainFileIsDir       error = getErr("main file should'nt be a directory"
  ErrSettingPWD          error = getErr("error while setting present working directory")
  ErrWatcherSetup        error = getErr("error while setting up watcher")
  ErrWalkingFS           error = getErr("error while walking file system")
)

// getErr() is a function that is in the repo; I haven't included it here
```

and these color sequences for flavorful logs.

```go
var (
	ResetColor    string = "\033[0m"
	GreenColor    string = "\033[0;32m"
	YellowColor   string = "\033[0;33m"
	BoldRedColor  string = "\033[1;31m"
	BoldCyanColor string = "\033[1;36m"
)
```

Here is the montre struct that I used:

```go
type Montre struct {
	args    []string
	pwd     string
	config  Config
	child   *exec.Cmd
	watcher *fsnotify.Watcher
	blocker chan struct{}
}
```

- I didn't really touch much on `args` and left it for future use cases.\
- `pwd` represents the present working directory, which is used for walking the root directory (default behavior).\
- `config` represents this structure:

  ```go
  type Config struct {
    MainFile   string   `json:"mainFile"`
    WatchExts  []string `json:"watchExtensions"`
    IgnoreDirs []string `json:"ignoreDirs"`
  }
  ```

  where all these fields are config fields that can be provided in `montre.json` config file and read by `montre`.
  Montre performs a file based watch rather that folder based watch (I was initially thinking of this approach), which enables it to have file based filtering of watch items.

- `child` holds reference to the child process that montre keeps restarting on change.
- `watcher` holds reference to `fsnotify`'s watcher instance, and
- `blocker` as the name suggests, it's for blocking the main context of the executing program. We need this because things such as accepting input from `stdin` for restarting or registering listening events are all `goroutines` as the main context can't be blocked, which is where Go's concurrency super powers kick in.

The `StartWatching` function starts the go routines for registering listening events, accepting input from the `stdin` and walking through all the files according to a given set of config or an extended set of config.

```go
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
```

and here is how `montre` reloads or restarts the child process. Here `m.child.Process.Kill()` is where I was facing the dreadful **nil pointer dereference issue** as the child process didn't exist due to some issues regarding clean exit of the process, as I was calling the `m.child.Wait()` on it, which was apparently throwing an error as the process was being killed while being awaited upon to release resources and exit on its own.

I was also at first trying to use the `m.child.StdoutPipe()` method for grabbing the `stdout` of the process and then using the `os.ReadAll()` method for reading and then calling `os.Stdout.Write()`, but I don't know why I was facing issues with it, so I switched it to the version of the code below where the **main process**'s `stdout` and `stderr` are the same for the child process `stdout` and `stderr`, kind of like piping the stdio of the child to the parent.

```go
func (m *Montre) quitChildProcess(errStr error) {
	if m.child != nil {
		err := m.child.Process.Kill()
		if err != nil {
			log.Fatalln(RedLog(errStr.Error() + " " + err.Error()))
		}
	}
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
```

The listening events are setup in the following manner, where we are listening to events of any kind and then reloading the server. On error, we are logging the error to the console.
The events case could be further expaned upon by for example filtering the process restart on specific events or some other functionality.

```go
func (m *Montre) initListenEvents() {
	for {
		select {
		case err, ok := <-m.watcher.Errors:
			if !ok {
				log.Fatalln(RedLog(err.Error()))
				return
			}
		case _, ok := <-m.watcher.Events:
			if !ok {
				return
			}
			m.reload()
		}
	}
}
```

I know that I briefly mentioned the implementation details, so do follow my [repo](https://github.com/ghoshRitesh12/montre) if interested.

## Conclusion

I learned a lot from building this small CLI tool, and since I am new to Go, I was stuck for hours trying to debug goroutine related issues and nil pointer dereferences. I would really like to hear some feedback on this article, the code or any ways that the code could be improved upon, as this is my first time writing a Go article.

\#Go \#Nodemon \#Montre \#Goroutine \#Concurrency
