package frontend

import (
  "github.com/jroimartin/gocui"
  "github.com/spf13/viper"
  "coco/backend"
  "fmt"
  "errors"
  "log"
  "os/exec"
  "os"
  "time"
)

var ErrNoConfig = errors.New("Config file not yet read!")
var ErrBadConfig = errors.New("Config file invalid!")

type Controller struct {
  Configuration *viper.Viper
  Gui *gocui.Gui

  stopChan chan bool
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

  err := os.MkdirAll(c.Configuration.GetString("Init.dir"), 0755)
  if err != nil {
    return err
  }
  if c.Configuration.GetString("Init.dir") !=
    c.Configuration.GetString("PeriodicCommand.dir") {
    err = os.MkdirAll(c.Configuration.GetString("PeriodicCommand.dir"), 0755)
    if err != nil {
      return err
    }
  }

  err = backend.RunInit(c.Configuration.GetString("Init.command"),
    c.Configuration.GetString("Init.dir"),
    c.Configuration.GetStringSlice("Init.args"),
    c.Configuration.GetStringSlice("Init.env"))
  if err != nil {
    return err
  }

  return nil
}

func (c *Controller) StartCommandLoop() error {
  if (c.Configuration == nil) {
    return ErrNoConfig
  }
  c.stopChan = make(chan bool)
  // Time case:
  if !c.Configuration.IsSet("RunOn.time") {
    return ErrBadConfig
  }
  go c.timeLoop()

  return nil
}

func (c *Controller) StopCommandLoop() {
  c.stopChan <- true
}

func (c *Controller) timeLoop() {
  // Read relevant config items:
  t_out := c.Configuration.GetFloat64("RunOn.time")
  command_str := c.Configuration.GetString("PeriodicCommand.command")
  command_args := c.Configuration.GetStringSlice("PeriodicCommand.args")
  command_dir := c.Configuration.GetString("PeriodicCommand.dir")
  command_env, err := backend.MakeEnvironment(
    c.Configuration.GetStringSlice("PeriodicCommand.env"))
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
      c := exec.Command(command_str, command_args...)
      c.Dir = command_dir
      c.Env = command_env
      com_chan <- c
    }

    output, err := (<-result)()
    if err != nil {
      log.Println(err)
    }

    c.ShowOutput(output)
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
