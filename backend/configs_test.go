package backend

import (
  "testing"
  "io"
  "os"
)

func testSetup(config_file string, t *testing.T) {
  src, err := os.Open("test_util/" + config_file)
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

func extTestSetup(config_file, path string, t *testing.T) {
  src, err := os.Open("test_util/" + config_file)
  if err != nil {
    t.Error(err)
  }
  dst, err := os.Create(path + "/" + "config.yaml")
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

func extTeardown(path string, t *testing.T) {
  err := os.Remove(path + "config.yaml")
  if err != nil {
    t.Error(err)
  }
}

func TestConfig(t *testing.T) {
  testSetup("test_config.yaml", t)
  defer testTeardown(t)

  viper, err := ReadConfig([]string{})
  if err != nil {
    t.Error(err)
  }

  viper.GetString("Init.command")
  if viper.GetString("Init.command") != "do-setup" {
    t.Error("Init command", viper.GetString("Init.command"), "expected do-setup")
  }

  exp_args := []string{"-f", "--longer-arg", "long_arg_val", "arg with space"}
  for i, val := range viper.GetStringSlice("Init.args") {
    if val != exp_args[i] {
      t.Error("Init args", viper.GetStringSlice("Init.args"), "expected", exp_args)
    }
  }

  if viper.GetString("Init.dir") != "/tmp/setup" {
    t.Error("Init dir", viper.GetString("Init.dir"), "expected /tmp/setup")
  }

  if viper.GetString("PeriodicCommand.command") != "real-command" {
    t.Error("PC", viper.GetString("PeriodicCommand.command"), "expected real-command")
  }

  if viper.GetString("PeriodicCommand.dir") != "/tmp/commands" {
    t.Error("PC dir", viper.GetString("PeriodicCommand.dir"), "expected /tmp/commands")
  }

  if viper.GetString("RunOn.mode") != "time" {
    t.Error("RunOn mode", viper.GetString("RunOn.mode"), "expected time")
  }

  if viper.GetFloat64("RunOn.time") != 1.66 {
    t.Error("RunOn time", viper.GetFloat64("RunOn.time"), "expected", 1.66)
  }

  if viper.GetString("RunOn.fs_root") != "/tmp/changeable" {
    t.Error("RunOn fs", viper.GetString("RunOn.fs_root"), "expected /tmp/changeable")
  }
}

func TestDefaults(t *testing.T) {
  testSetup("defaults_config.yaml", t)
  defer testTeardown(t)

  viper, err := ReadConfig([]string{})
  if err != nil {
    t.Error(err)
  }

  if viper.GetString("Init.dir") != "/tmp/coco" {
    t.Fail()
  }

  if viper.GetString("PeriodicCommand.dir") != "/tmp/coco" {
    t.Fail()
  }
}

func TestBadConfig(t *testing.T) {
  testSetup("bad_config.yaml", t)
  defer testTeardown(t)

  _, err := ReadConfig([]string{})
  t.Log(err)
  if err == nil {
    t.Fail()
  }
}

func TestAddPath(t *testing.T) {
  extTestSetup("test_config.yaml", "/tmp/", t)
  defer extTeardown("/tmp/", t)

  viper, err := ReadConfig([]string{"/tmp"})

  if err != nil {
    t.Error(err)
  }

  viper.GetString("Init.command")
  if viper.GetString("Init.command") != "do-setup" {
    t.Error("Init command", viper.GetString("Init.command"), "expected do-setup")
  }
}

func TestNoConfig(t *testing.T) {
  _, err := ReadConfig([]string{})
  t.Log(err)
  if err == nil {
    t.Fail()
  }
}

func TestCommandModeTime(t *testing.T) {
  testSetup("t_conf_1.yaml", t)
  defer testTeardown(t)

  viper, err := ReadConfig([]string{})
  if err != nil {
    t.Error(err)
  }

  if GetCommandMode(viper) != TimeMode {
    t.Error("Expected time mode")
  }

  testSetup("t_conf_2.yaml", t)

  viper, err = ReadConfig([]string{})
  if err != nil {
    t.Error(err)
  }

  if GetCommandMode(viper) != TimeMode {
    t.Error("Expected time mode")
  }

  testSetup("t_conf_3.yaml", t)

  viper, err = ReadConfig([]string{})
  if err != nil {
    t.Error(err)
  }

  if GetCommandMode(viper) != Invalid {
    t.Error("Expected invalid mode")
  }
}

func TestCommandModeFS(t *testing.T) {
  testSetup("fs_conf_1.yaml", t)
  defer testTeardown(t)

  viper, err := ReadConfig([]string{})
  if err != nil {
    t.Error(err)
  }

  if GetCommandMode(viper) != FSMode {
    t.Error("Expected fs mode")
  }

  testSetup("fs_conf_2.yaml", t)

  viper, err = ReadConfig([]string{})
  if err != nil {
    t.Error(err)
  }

  if GetCommandMode(viper) != FSMode {
    t.Error("Expected fs mode, got", int(GetCommandMode(viper)))
  }

  testSetup("fs_conf_3.yaml", t)

  viper, err = ReadConfig([]string{})
  if err != nil {
    t.Error(err)
  }

  if GetCommandMode(viper) != Invalid {
    t.Error("Expected invalid mode")
  }
}

func TestCommandModeInvalid(t *testing.T) {
  testSetup("inv_conf_1.yaml", t)
  defer testTeardown(t)

  viper, err := ReadConfig([]string{})
  if err != nil {
    t.Error(err)
  }

  if GetCommandMode(viper) != Invalid {
    t.Error("Expected inavlid mode")
  }

  testSetup("inv_conf_2.yaml", t)

  viper, err = ReadConfig([]string{})
  if err != nil {
    t.Error(err)
  }

  if GetCommandMode(viper) != Invalid {
    t.Error("Expected inavlid mode")
  }
}

func TestPeriodicCommand(t *testing.T) {
  testSetup("test_config.yaml", t)
  defer testTeardown(t)

  viper, err := ReadConfig([]string{})
  if err != nil {
    t.Error(err)
  }

  com, err := ReadPeriodicCommand(viper)
  if err != nil {
    t.Error(err)
  }

  if com.Name != "real-command" || com.Dir != "/tmp/commands" {
    t.Error("Created incorrect command")
  }
}
