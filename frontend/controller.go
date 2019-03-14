package frontend

import (
  "github.com/jroimartin/gocui"
  "github.com/spf13/viper"
  "github.com/fsnotify/fsnotify"
  "github.com/MikeKneeB/coco/backend"
  "fmt"
  "errors"
  "log"
  "os/exec"
  "os"
  "time"
  "path/filepath"
  "strings"
)

var ErrNoConfig = errors.New("Config file not yet read!")
var ErrBadConfig = errors.New("Config file invalid!")

type Controller struct {
  Configuration *viper.Viper
  Gui *gocui.Gui

  stopChan chan bool
}

func NewController() *Controller {
  c := new(Controller)

  c.stopChan = make(chan bool)

  return c
}

func (c *Controller) AddGui(g *gocui.Gui) error {
  c.Gui = g
  g.SetManager(c)

  // Key Binds?
  err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, c.quit)
  if err != nil {
    return err
	}

  return nil
}

func (c *Controller) ReadConfig() error {
  var err error
  c.Configuration, err = backend.ReadConfig([]string{})
  if err != nil {
    return err
  }

  return nil
}

func (c *Controller) quit(g *gocui.Gui, v *gocui.View) error {
  c.stopChan <- true
  return gocui.ErrQuit
}

func (c *Controller) Layout(g *gocui.Gui) error {
  max_x, max_y := g.Size()
  _, err := g.SetView("normal", 0, 0, max_x - 1, max_y - 1)
  if err != nil && err != gocui.ErrUnknownView {
    return err
  }
  return nil
}

func (c *Controller) Init() error {
  if (c.Configuration == nil) {
    return ErrNoConfig
  }

  err := os.MkdirAll(c.Configuration.GetString("PeriodicCommand.dir"), 0755)
  if err != nil {
    return err
  }
  if c.Configuration.GetString("Init.dir") !=
    c.Configuration.GetString("PeriodicCommand.dir") {
    err = os.MkdirAll(c.Configuration.GetString("Init.dir"), 0755)
    if err != nil {
      return err
    }
  }

  if c.Configuration.IsSet("Init.Command") {
    err = backend.RunInit(c.Configuration.GetString("Init.command"),
      c.Configuration.GetString("Init.dir"),
      c.Configuration.GetStringSlice("Init.args"),
      c.Configuration.GetStringSlice("Init.env"))
    if err != nil {
      return err
    }
  }

  return nil
}

// TODO: Move command loop type to backend
func (c *Controller) StartCommandLoop() error {
  if (c.Configuration == nil) {
    return ErrNoConfig
  }

  switch mode := backend.GetCommandMode(c.Configuration) ; mode {
  case backend.TimeMode:
    go c.timeLoop()
  case backend.FSMode:
    go c.fsLoop()
  case backend.SignalMode:
  default:
    return ErrBadConfig
  }
  return nil
}

func (c *Controller) StopCommandLoop() {
  c.stopChan <- true
}


func (c *Controller) timeLoop() {
  // Read relevant config items:
  t_out := c.Configuration.GetFloat64("RunOn.time")
  com, err := backend.ReadPeriodicCommand(c.Configuration)
  if err != nil {
    // Call some kind of error report?
    log.Println(err)
    return
  }
  com_chan := make(chan *exec.Cmd)
  quit_chan := make(chan bool)
  result := make(chan backend.RoutineOut)

  go backend.ContinualRoutine(result, quit_chan, com_chan)

  for {

    select {
    case _ = <-c.stopChan:
      quit_chan <- true
      return
    case <-time.After(time.Duration(t_out * 1000000000 ) * time.Nanosecond):
      com_chan <- com.MakeRunnable()
    }

    output, err := (<-result)()
    if err != nil {
      _, ok := err.(*exec.ExitError)
      if !ok {
        log.Println(err)
      }
    }

    c.ShowOutput(output)
  }
}

func (c *Controller) fsLoop() {
  log.Println("Enter fs loop")
  com, err := backend.ReadPeriodicCommand(c.Configuration)
  if err != nil {
    log.Println(err)
    return
  }

  root := c.Configuration.GetString("RunOn.fs_root")
  exts := c.Configuration.GetStringSlice("RunOn.fs_extensions")

  check := (len(exts) != 0)

  com_chan := make(chan *exec.Cmd)
  quit_chan := make(chan bool)
  result := make(chan backend.RoutineOut)

  log.Println("Starting continual routine")
  go backend.ContinualRoutine(result, quit_chan, com_chan)

  log.Println("Make watcher")
  watcher, err := fsnotify.NewWatcher()
  if err != nil {
    log.Println(err)
    return
  }
  defer watcher.Close()

  log.Println("Add files from: ", root)
  err = filepath.Walk(root,
    func(path string, info os.FileInfo, err error) error {
      log.Println("walking", path, info, err)
      if err == nil && info.IsDir() {
        log.Println("Adding: ", path)
        watcher.Add(path)
      }
      return nil
    })

  log.Println("Done adding")
  send_comm := false

  com_chan <- com.MakeRunnable()

  output, err := (<-result)()
  if err != nil {
    _, ok := err.(*exec.ExitError)
    if !ok {
      log.Println(err)
    } else {
      c.ShowOutput(output)
    }
  } else {
    c.ShowAllGood()
  }

  for {
    send_comm = false
    select {
    case _ = <-c.stopChan:
      quit_chan <- true
      return
    case event, ok := <-watcher.Events:
      if !ok {
        quit_chan <- true
        return
      }
      if !(event.Op == fsnotify.Chmod) {
        if check {
          for _, v := range exts {
            if strings.HasSuffix(event.Name, "." + v) {
              send_comm = true
            }
          }
        } else {
          send_comm = true
        }
      }
    case error, ok := <-watcher.Errors:
      if ok {
        log.Println(error)
      }
      quit_chan <- true
      return
    }

    if send_comm {
      com_chan <- com.MakeRunnable()

      output, err := (<-result)()
      if err != nil {
        // Instead should parse using outp & ec
        _, ok := err.(*exec.ExitError)
        if !ok {
          log.Println(err)
        } else {
          c.ShowOutput(output)
        }
      } else {
        c.ShowAllGood()
      }
    }
  }
}

func (c *Controller) ShowOutput(output string) {
  c.Gui.Update(func(g *gocui.Gui) error {
    v, err := g.View("normal")
    if err != nil {
      return err
    }
    v.Clear()
    fmt.Fprint(v, output)
    return nil
  })
}

func (c *Controller) ShowAllGood() {
  c.Gui.Update(func(g *gocui.Gui) error {
    v, err := g.View("normal")
    if err != nil {
      return err
    }
    v.Clear()
    fmt.Fprint(v, "Looks good!")
    return nil
  })
}
