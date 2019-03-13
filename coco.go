package main

import (
  "github.com/jroimartin/gocui"
  "coco/frontend"
  "log"
)

func main() {
  g, err := frontend.Begin()
  if err != nil {
    log.Panicln(err)
  }
  defer g.Close()

  err = g.MainLoop()
  if err != nil && err != gocui.ErrQuit {
    log.Panicln(err)
  }
}
