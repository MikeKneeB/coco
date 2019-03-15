package frontend

import (
  "github.com/MikeKneeB/coco/backend"
  "errors"
  "os/exec"
  "time"
  "github.com/fsnotify/fsnotify"
  "path/filepath"
  "os"
  "strings"
)

var ErrNoOuputFn error = errors.New("Output Function not defined!")

type OutputFunction func(out string, ret_code int)
type LogFunction func(items ...interface{})
type OpFunction func(op string)

type RunnerSignal int

const (
  Quit RunnerSignal = iota
  ForceUpdate
)

// Common composition types
// Functions compositor - this is exposed because the controller needs to
// populate it
type runnerFuncs struct {
  outputFunc OutputFunction
  logFunc LogFunction
  opFunc OpFunction
}

func NewRunnerFuncs(of OutputFunction, lf LogFunction, op OpFunction) runnerFuncs {
  return runnerFuncs{of, lf, op}
}

// Channels compositor
type runnerChannels struct {
  sigChan chan RunnerSignal
  comChan chan *exec.Cmd
  resChan chan backend.RoutineOut
  quitChan chan bool
}

func newRunnerChannels() runnerChannels {
  return runnerChannels{make(chan RunnerSignal), make(chan *exec.Cmd),
    make(chan backend.RoutineOut), make(chan bool)}
}

type Runner interface {
  Start() error
  Signal(sig RunnerSignal)
}

type TimeRunner struct {
  // Actual struct data
  timeOut time.Duration
  command backend.CommandDef
  // Callbacks
  runnerFuncs
  // Channels
  runnerChannels
}

func NewTimeRunner(to float64, c backend.CommandDef, rf runnerFuncs) *TimeRunner {
  r := new(TimeRunner)
  r.timeOut = time.Duration(to * 1000000000) * time.Nanosecond
  r.command = c
  r.runnerFuncs = rf
  r.runnerChannels = newRunnerChannels()
  return r
}

func (r *TimeRunner) Start() error {
  if r.outputFunc == nil {
    return ErrNoOuputFn
  }

  go r.loop()

  return nil
}

func (r *TimeRunner) loop() {

  go backend.ContinualRoutine(r.resChan, r.quitChan, r.comChan)

  for {
    // Part A - wait for signals (or t-out)
    r.wait()
    // Part B - send result back
    r.send()
    // End loop
  }
}

func (r *TimeRunner) wait() {
  select {
  case sig := <- r.sigChan:
    if sig == Quit {
      r.quitChan <- true
      return
    } else if sig == ForceUpdate {
      r.comChan <- r.command.MakeRunnable()
    }
  case <- time.After(r.timeOut):
    r.comChan <- r.command.MakeRunnable()
  }
}

func (r *TimeRunner) send() {
  r.logFunc("Run command: ", r.command)
  r.opFunc("EXECUTING")
  output, err := (<- r.resChan)()
  if err != nil {
    exit_err, ok := err.(*exec.ExitError)
    if !ok {
      r.logFunc(err)
    } else {
      r.logFunc("Command exited: ", exit_err.ExitCode())
      r.outputFunc(output, exit_err.ExitCode())
    }
  } else {
    r.logFunc("Command Exited: ", 0)
    r.outputFunc(output, 0)
  }
  r.opFunc("IDLE")
}

func (r *TimeRunner) Signal(sig RunnerSignal) {
  r.sigChan <- sig
}

type FSRunner struct {
  // Actual data
  root string
  exts []string
  command backend.CommandDef
  // Callbacks
  runnerFuncs
  // Signals
  runnerChannels
  // FS watcher;
  watcher *fsnotify.Watcher
}

func NewFSRunner(root string, exts []string, c backend.CommandDef,
  rf runnerFuncs) (*FSRunner, error) {
  r := new(FSRunner)
  r.root = root
  r.exts = exts
  r.command = c
  r.runnerFuncs = rf
  r.runnerChannels = newRunnerChannels()
  var err error
  r.watcher, err = fsnotify.NewWatcher()
  if err != nil {
    return nil, err
  }
  return r, nil
}

func (r *FSRunner) Start() error {
  if r.outputFunc == nil {
    return ErrNoOuputFn
  }

  err := r.addWatchedFolders()
  if err != nil {
    return err
  }

  go r.loop()

  return nil
}

type fsStatus int

const (
  fsSend fsStatus = iota
  fsContinue
  fsQuit
)

func (r *FSRunner) loop() {
  go backend.ContinualRoutine(r.resChan, r.quitChan, r.comChan)

  check := (len(r.exts) != 0)
  send_command := fsSend

  for {
    // Part A - send command and respond
    if send_command == fsSend {
      r.send()
    } else if send_command == fsQuit {
      return
    }

    // Part B - wait for signals
    send_command = r.wait(check)
  }
}

func (r *FSRunner) wait(check bool) fsStatus {
  select {
  case sig := <- r.sigChan:
    if sig == Quit {
      r.quitChan <- true
      return fsQuit
    } else if sig == ForceUpdate {
      return fsSend
    }
  case event, ok := <- r.watcher.Events:
    if !ok {
      r.quitChan <- true
      return fsContinue
    }
    if !(event.Op == fsnotify.Chmod) {
      if check {
        return r.checkUpdate(event)
      } else {
        return fsSend
      }
    }
  case error, ok := <- r.watcher.Errors:
    if ok {
      r.logFunc(error)
    }
    r.quitChan <- true
    return fsQuit
  }
  return fsQuit
}

func (r *FSRunner) send() {
  r.logFunc("Run command: ", r.command)
  r.opFunc("EXECUTING")
  r.comChan <- r.command.MakeRunnable()

  output, err := (<- r.resChan)()
  if err != nil {
    exit_err, ok := err.(*exec.ExitError)
    if !ok {
      r.logFunc(err)
    } else {
      r.logFunc("Command exited: ", exit_err.ExitCode())
      r.outputFunc(output, exit_err.ExitCode())
    }
  } else {
    r.logFunc("Command exited: ", 0)
    r.outputFunc(output, 0)
  }
  r.opFunc("IDLE")
}

func (r *FSRunner) checkUpdate(e fsnotify.Event) fsStatus {
  for _, v := range r.exts {
    if strings.HasSuffix(e.Name, "." + v) {
      return fsSend
    }
  }
  return fsContinue
}

func (r *FSRunner) addWatchedFolders() error {
  return filepath.Walk(r.root,
    func(path string, info os.FileInfo, err error) error {
      if err == nil && info.IsDir() {
        r.watcher.Add(path)
      }
      return nil
    })
}

func (r *FSRunner) Signal(sig RunnerSignal) {
  r.sigChan <- sig
}
