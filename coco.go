package main

import (
  "github.com/jroimartin/gocui"
  "coco/frontend"
  "log"
  "os"
)

func main() {
  f, err := os.Create("coco_log.txt")
  if err != nil {
    log.Panicln(err)
  }
  defer f.Close()
  log.SetOutput(f)
  log.Println("WHAT")
  c, err := frontend.NewController()
  log.Println("HMM")
  if err != nil {
    log.Panicln(err)
  }
  defer c.Gui.Close()

  err = c.StartCommandLoop()
  log.Println("EH?")
  if err != nil {
    log.Panicln(err)
  }

  err = c.Gui.MainLoop()
  log.Println("FUCK")
  if err != nil && err != gocui.ErrQuit {
    log.Panicln(err)
  }
}
