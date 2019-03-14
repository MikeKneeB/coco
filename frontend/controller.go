package frontend

import (
  "github.com/jroimartin/gocui"
  "github.com/spf13/viper"
  "github.com/MikeKneeB/coco/backend"
  "fmt"
  "errors"
  "os"
  "log"
  "time"
)

var ErrNoConfig = errors.New("Config file not yet read!")
var ErrBadConfig = errors.New("Config file invalid!")

type Controller struct {
  Configuration *viper.Viper
  Gui *gocui.Gui

  runner Runner
}

func NewController() *Controller {
  c := new(Controller)

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
  c.StopCommandLoop()
  return gocui.ErrQuit
}

func (c *Controller) Layout(g *gocui.Gui) error {
  max_x, max_y := g.Size()

  v, err := g.SetView("normal", 0, 0, max_x - 1, max_y - 9)
  if err == gocui.ErrUnknownView {
    v.Wrap = true
  } else if err != nil {
    return err
  }

  v, err = g.SetView("log", 0, max_y - 8, max_x - 1, max_y - 1)
  if err == gocui.ErrUnknownView {
    v.Wrap = true
    v.Autoscroll = true
    v.Title = "Log"
  } else if err != nil {
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

  com, err := backend.ReadPeriodicCommand(c.Configuration)
  if err != nil {
    return err
  }

  switch mode := backend.GetCommandMode(c.Configuration) ; mode {
  case backend.TimeMode:
    c.runner = NewTimeRunner(c.Configuration.GetFloat64("RunOn.time"), com,
    c.ShowOutput, c.Log)
  case backend.FSMode:
    c.runner, err = NewFSRunner(c.Configuration.GetString("RunOn.fs_root"),
    c.Configuration.GetStringSlice("RunOn.fs_extensions"), com, c.ShowOutput,
    c.Log)
    if err != nil {
      return err
    }
  case backend.SignalMode:
  default:
    return ErrBadConfig
  }

  return c.runner.Start()
}

func (c *Controller) StopCommandLoop() {
  log.Println("Call stop command")
  c.runner.Signal(Quit)
  log.Println("Done stopping command")
}

func (c *Controller) ShowOutput(output string, return_code int) {
  c.Gui.Update(func(g *gocui.Gui) error {
    v, err := g.View("normal")
    if err != nil {
      return err
    }
    v.Clear()
    if return_code == 0 {
      fmt.Fprint(v, "Looks good!")
    } else {
      fmt.Fprint(v, output)
    }
    return nil
  })
}

func (c *Controller) Log(items ...interface{}) {
  c.Gui.Update(func(g *gocui.Gui) error {
    v, err := g.View("log")
    if err != nil {
      return err
    }
    fmt.Fprint(v, "\n" + time.Now().Format("2006/01/02 15:04:05 "))
    for _, log_item := range items {
      fmt.Fprint(v, log_item, "")
    }
    return nil
  })
}
