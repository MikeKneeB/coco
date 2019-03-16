package main

import (
  "github.com/jroimartin/gocui"
  "github.com/MikeKneeB/coco/frontend"
  "log"
)

func main() {
  c := frontend.NewController()
  g, err := gocui.NewGui(gocui.OutputNormal)
  if err != nil {
    log.Panicln(err)
  }
  defer g.Close()

  // Add created gui to the controller!
  c.AddGui(g)

  // Read config file
  err = c.ReadConfig()
  if err != nil {
    log.Panicln(err)
  }

  // Make necessary dirs, run Init command
  err = c.Init()
  if err != nil {
    log.Panicln(err)
  }

  // Begin processing loop
  err = c.StartCommandLoop()
  if err != nil {
    log.Panicln(err)
  }

  // Begin gui's main loop - blocks until quit signal sent
  err = c.Gui.MainLoop()
  if err != nil && err != gocui.ErrQuit {
    log.Panicln(err)
  }
}
