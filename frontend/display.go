package frontend

import (
  "github.com/jroimartin/gocui"
  "coco/backend"
)

func Begin() (*Controller, error) {
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

  err = g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit)
  if err != nil {
    return nil, err
	}

  return c, nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
  return gocui.ErrQuit
}
