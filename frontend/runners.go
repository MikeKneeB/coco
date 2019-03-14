package frontend

import (
  "github.com/MikeKneeB/coco/backend"
  "errors"
  "os/exec"
  "time"
  "log"
  "github.com/fsnotify/fsnotify"
  "path/filepath"
  "os"
)

var ErrNoOuputFn error = errors.New("Output Function not defined!")

type RunnerSignal int

const (
  Quit RunnerSignal = iota
  ForceUpdate
)

type OutputFunction func(out string, ret_code int)

type Runner interface {
  Start() error
  Signal(sig RunnerSignal)
}

type TimeRunner struct {
  timeOut time.Duration
  command backend.CommandDef
  outputFunc OutputFunction
  sigChan chan RunnerSignal
}

func NewTimeRunner(to float64, c backend.CommandDef,
  of OutputFunction) *TimeRunner {
  r := new(TimeRunner)
  r.timeOut = time.Duration(to * 1000000000) * time.Nanosecond
  r.command = c
  r.outputFunc = of
  r.sigChan = make(chan RunnerSignal)
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
  com_chan := make(chan *exec.Cmd)
  res_chan := make(chan backend.RoutineOut)
  quit_chan := make(chan bool)

  go backend.ContinualRoutine(res_chan, quit_chan, com_chan)

  for {
    // Part A - wait for signals (or t-out)
    select {
    case sig := <- r.sigChan:
      if sig == Quit {
        quit_chan <- true
        return
      } else if sig == ForceUpdate {
        com_chan <- r.command.MakeRunnable()
      }
    case <- time.After(r.timeOut):
      com_chan <- r.command.MakeRunnable()
    }

    // Part B - send result back
    output, err := (<-res_chan)()
    if err != nil {
      exit_err, ok := err.(*exec.ExitError)
      if !ok {
        // Log error - use something smarter than current log tho
        log.Println(err)
      } else {
        r.outputFunc(output, exit_err.ExitCode())
      }
    } else {
      r.outputFunc(output, 0)
    }

    // End loop
  }
}

func (r *TimeRunner) Signal(sig RunnerSignal) {
  r.sigChan <- sig
}

type FSRunner struct {
  root string
  exts []string
  outputFunc OutputFunction
  sigChan chan RunnerSignal
  watcher *fsnotify.Watcher
}

func NewFSRunner(root string, exts []string, output OutputFunction) (*FSRunner,
  error) {
  r := new(FSRunner)
  r.root = root
  r.exts = exts
  r.outputFunc = output
  r.sigChan = make(chan RunnerSignal)
  var err error
  r.watcher, err = fsnotify.NewWatcher()
  if err != nil {
    return nil, err
  }
  return r, nil
}

func (r *FSRunner) Close() {
  r.watcher.Close()
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

func (r *FSRunner) loop() {
  com_chan := make(chan *exec.Cmd)
  res_chan := make(chan backend.RoutineOut)
  quit_chan := make(chan bool)

  go backend.ContinualRoutine(res_chan, quit_chan, com_chan)

  send_command := true

  for {
    // Part A - send command and respond
    if send_command {
      
    }
    select {
    case sig := <- r.sigChan:
      if sig == Quit {
        quit_chan <- true
        return
      } else if sig == ForceUpdate {

      }
    }
    // Part B - send results
  }
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
