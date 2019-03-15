package backend

import (
  "github.com/spf13/viper"
  "fmt"
)

var required_confs = [...]string {"PeriodicCommand.command"}
var default_confs = map[string]interface{} {"Init.dir": "/tmp/coco",
                                           "PeriodicCommand.dir": "/tmp/coco"}

func checkRequired(viper *viper.Viper) error {
  for _, val := range required_confs {
    if !viper.IsSet(val) {
      return fmt.Errorf("Did not find required field %s in config file %s", val,
                        viper.ConfigFileUsed())
    }
  }
  return nil
}

func setDefaults(viper *viper.Viper) {
  for key, val := range default_confs {
    viper.SetDefault(key, val)
  }
}

func ReadConfig(paths []string) (*viper.Viper, error) {
  viper := viper.New()
  viper.AddConfigPath(".")
  for _, v := range paths {
    viper.AddConfigPath(v)
  }
  viper.SetConfigName("config")
  setDefaults(viper)
  err := viper.ReadInConfig()
  if err != nil {
    return nil, err
  }
  err = checkRequired(viper)
  if err != nil {
    return nil, err
  }
  return viper, nil
}

type CommandMode int

const (
  TimeMode CommandMode = iota
  FSMode
  SignalMode
  Invalid
)

func GetCommandMode(viper *viper.Viper) CommandMode {
  if viper.IsSet("RunOn.mode") {
    switch mode := viper.GetString("RunOn.mode"); mode {
    case "time":
      if !viper.IsSet("RunOn.time") {
        return Invalid
      }
      return TimeMode
    case "fs":
      if !viper.IsSet("RunOn.fs_root") {
        return Invalid
      }
      return FSMode
    default:
      return Invalid
    }
  } else {
    if viper.IsSet("RunOn.time") {
      return TimeMode
    } else if viper.IsSet("RunOn.fs_root") {
      return FSMode
    } else {
      return Invalid
    }
  }
}

func ReadPeriodicCommand(viper *viper.Viper) (CommandDef, error) {
  com := CommandDef{Name: viper.GetString("PeriodicCommand.command"),
  Dir: viper.GetString("PeriodicCommand.dir"),
  Args: viper.GetStringSlice("PeriodicCommand.args"), Env: []string{}}
  var err error
  com.Env, err = MakeEnvironment(
    viper.GetStringSlice("PeriodicCommand.env"))
  return com, err
}

/*
Init:
  command : cmake
  args:
    - -Gninja
    - /home/mikek/whatever...
  dir : /tmp/build

GeneralCommand:
  command : ninja
  dir : /tmp/build
*/
