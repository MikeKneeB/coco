package backend

import (
  "github.com/spf13/viper"
  "fmt"
)

var required_confs = [...]string {"PeriodicCommand.command"}
var default_confs = map[string]interface{} {"Init.dir": "/tmp/coco",
                                           "PeriodicCommand.dir": "/tmp/coco",
                                           "RunOn.time": 1.0}

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

func ReadConfig() (*viper.Viper, error) {
  viper := viper.New()
  viper.AddConfigPath(".")
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
