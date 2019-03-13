package frontend

import (
  "github.com/jroimartin/gocui"
  "coco/backend"
)

func Begin() (*gocui.Gui, error) {
  g, err := gocui.NewGui(gocui.OutputNormal)
  if err != nil {
    return nil, err
  }
  c := new(Controller)
  c.Configuration, err = backend.ReadConfig()
  c.Gui = g
  g.SetManager(c)

  err = c.Init()
  if err != nil {
    return nil, err
  }
  c.StartCommandLoop()

  err = g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit)
  if err != nil {
    return nil, err
	}

  return g, nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
  return gocui.ErrQuit
}
