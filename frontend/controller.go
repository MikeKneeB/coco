package frontend

import (
  "github.com/jroimartin/gocui"
  "github.com/spf13/viper"
  "github.com/MikeKneeB/coco/backend"
  "fmt"
  "errors"
  "os"
  "time"
  "strings"
)

var ErrNoConfig = errors.New("Config file not yet read!")
var ErrBadConfig = errors.New("Config file invalid!")

type Controller struct {
  Configuration *viper.Viper
  Gui *gocui.Gui

  runner Runner

  operation string
  logY int
}

func NewController() *Controller {
  c := new(Controller)

  c.operation = "IDLE"
  c.logY = 1
  return c
}

func (c *Controller) AddGui(g *gocui.Gui) error {
  c.Gui = g
  g.SetManager(c)

  // Key Binds
  err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, c.quit)
  if err != nil {
    return err
	}

  err = g.SetKeybinding("", 'u', gocui.ModNone, c.forceUpdate)
  if err != nil {
    return err
  }

  err = g.SetKeybinding("", 'l', gocui.ModNone, c.toggleLog)
  if err != nil {
    return err
  }

  return nil
}

func (c *Controller) quit(g *gocui.Gui, v *gocui.View) error {
  c.StopCommandLoop()
  return gocui.ErrQuit
}

func (c *Controller) forceUpdate(g *gocui.Gui, v *gocui.View) error {
  c.runner.Signal(ForceUpdate)
  return nil
}

func (c *Controller) toggleLog(g *gocui.Gui, v *gocui.View) error {
  c.logY = (c.logY + 1) % 2
  return nil
}

func (c *Controller) ReadConfig() error {
  var err error
  c.Log("Reading config file...")
  c.Configuration, err = backend.ReadConfig([]string{})
  if err != nil {
    c.Log("Error reading config file: ", err)
    return err
  }

  c.Log("Successfully got configuration.")
  return nil
}

func (c *Controller) Layout(g *gocui.Gui) error {
  max_x, max_y := g.Size()

  v, err := g.SetView("operation", 0, -1, max_x - 1, 2)
  if err == gocui.ErrUnknownView {
    v.Frame = false
    v.FgColor = gocui.AttrBold
  } else if err != nil {
    return err
  }
  v.Clear()
  fmt.Fprint(v, c.operation)

  v, err = g.SetView("log", 0, max_y - 8, max_x - 1, max_y - 1)
  if err == gocui.ErrUnknownView {
    v.Wrap = true
    v.Autoscroll = true
    v.Title = "Log"
  } else if err != nil {
    return err
  }

  v, err = g.SetView("normal", 0, 1, max_x - 1, max_y - (c.logY * 8 + 1))
  if err == gocui.ErrUnknownView {
    v.Wrap = true
  } else if err != nil {
    return err
  }

  return nil
}

func (c *Controller) Init() error {
  if (c.Configuration == nil) {
    return ErrNoConfig
  }

  pcDir := c.Configuration.GetString("PeriodicCommand.dir")
  c.Log("Creating configured periodic command directory:", pcDir)
  err := os.MkdirAll(pcDir, 0755)
  if err != nil {
    c.Log(err)
    return err
  }
  icDir := c.Configuration.GetString("Init.dir")
  if icDir != pcDir {
    c.Log("Creating configured init command directory:", icDir)
    err = os.MkdirAll(icDir, 0755)
    if err != nil {
      c.Log(err)
      return err
    }
  }

  if c.Configuration.IsSet("Init.Command") {
    com := c.Configuration.GetString("Init.command")
    args := c.Configuration.GetStringSlice("Init.args")
    dir := c.Configuration.GetString("Init.dir")
    env := c.Configuration.GetStringSlice("Init.env")
    c.Log("Running init command:", com, args)
    err = backend.RunInit(com, dir, args, env)
    if err != nil {
      return err
    }
  }

  return nil
}

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
    NewRunnerFuncs(c.ShowOutput, c.Log, c.UpdateOperation))
  case backend.FSMode:
    c.runner, err = NewFSRunner(c.Configuration.GetString("RunOn.fs_root"),
    c.Configuration.GetStringSlice("RunOn.fs_extensions"), com,
    NewRunnerFuncs( c.ShowOutput, c.Log, c.UpdateOperation))
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
  c.Log("Call stop command")
  c.runner.Signal(Quit)
}

func (c *Controller) ShowOutput(output string, return_code int) {
  var cls []backend.CompileLine
  if return_code != 0 {
    cls = backend.Parse(output)
  }

  c.Gui.Update(func(g *gocui.Gui) error {
    v, err := g.View("normal")
    if err != nil {
      return err
    }
    v.Clear()
    if return_code == 0 {
      fmt.Fprint(v, "\033[32;1mLooks good!\033[0m")
    } else {
      for _, line := range cls {
        fmt.Fprintf(v, "%s\n%d\n%s\n", line.FileName, line.Line,
          strings.TrimSpace(line.Message))
      }
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

func (c *Controller) UpdateOperation(op string) {
  c.operation = op
  c.Gui.Update(func(g *gocui.Gui) error {
    return nil
  })
}
