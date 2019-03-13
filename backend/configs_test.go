package backend

import (
  "testing"
  //"github.com/spf13/viper"
  "io"
  "os"
)

func testSetup(config_file string, t *testing.T) {
  src, err := os.Open(config_file)
  if err != nil {
    t.Error(err)
  }
  dst, err := os.Create("config.yaml")
  if err != nil {
    t.Error(err)
  }
  _, err = io.Copy(dst, src)
  if err != nil {
    t.Error(err)
  }
}

func testTeardown(t *testing.T) {
  err := os.Remove("config.yaml")
  if err != nil {
    t.Error(err)
  }
}

func TestConfig(t *testing.T) {
  testSetup("test_config.yaml", t)
  defer testTeardown(t)

  viper, err := ReadConfig()
  if err != nil {
    t.Error(err)
  }

  viper.GetString("Init.command")
  if viper.GetString("Init.command") != "do-setup" {
    t.Fail()
  }

  exp_args := []string{"-f", "--longer-arg", "long_arg_val", "arg with space"}
  for i, val := range viper.GetStringSlice("Init.args") {
    if val != exp_args[i] {
      t.Fail()
    }
  }

  if viper.GetString("Init.dir") != "/tmp/setup" {
    t.Fail()
  }

  if viper.GetString("PeriodicCommand.command") != "real-command" {
    t.Fail()
  }

  if viper.GetString("PeriodicCommand.dir") != "/tmp/commands" {
    t.Fail()
  }

  if viper.GetString("RunOn.mode") != "time" {
    t.Fail()
  }

  if viper.GetFloat64("RunOn.time") != 1.66 {
    t.Fail()
  }

  if viper.GetString("RunOn.fs_change") != "/tmp/changeable" {
    t.Fail()
  }
}

func TestDefaults(t *testing.T) {
  testSetup("defaults_config.yaml", t)
  defer testTeardown(t)

  viper, err := ReadConfig()
  if err != nil {
    t.Error(err)
  }

  if viper.GetString("Init.dir") != "/tmp/coco" {
    t.Fail()
  }

  if viper.GetString("PeriodicCommand.dir") != "/tmp/coco" {
    t.Fail()
  }

  if viper.GetFloat64("RunOn.time") != 1.0 {
    t.Fail()
  }
}

func TestBadConfig(t *testing.T) {
  testSetup("bad_config.yaml", t)
  defer testTeardown(t)

  _, err := ReadConfig()
  t.Log(err)
  if err == nil {
    t.Fail()
  }
}
