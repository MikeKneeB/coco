package backend

import (
  "os/exec"
  "os"
  "strings"
  "errors"
  "fmt"
)

var ErrRoutineQuit = errors.New("Quit Continual Routine")

type RoutineOut func()(string, error)

func RunRoutine(result chan<- RoutineOut, command *exec.Cmd) {
  byte_output, err := command.CombinedOutput()
  result <- (func()(string, error){ return string(byte_output), err })
}

func ContinualRoutine(result chan<- RoutineOut, quit <-chan bool,
    command <-chan *exec.Cmd) {
  for {
    select {
    case _ = <-quit:
      result <- (func()(string, error){ return "", ErrRoutineQuit })
      return
    case c := <-command:
      RunRoutine(result, c)
    }
    /*q := <-quit
    if q {
      result <- (func()(string, error){ return "", ErrRoutineQuit })
      return
    } else {
      RunRoutine(result, command)
    }*/
  }
}

func RunInit(command, dir string, args, config_env []string) error {
  c := exec.Command(command, args...)
  envs, err := MakeEnvironment(config_env)
  if err != nil {
    return err
  }
  c.Env = envs
  c.Dir = dir
  return c.Run()
}

func MakeEnvironment(add []string) ([]string, error) {
  env := os.Environ()
  for _, v := range add {
    kv_slice := strings.Split(v, "=")
    if len(kv_slice) != 2 {
      /* TODO: Also check validity of key, cannot contain spaces or $ etc. */
      return nil, fmt.Errorf("Poorly formatted environment item: %s", v)
    }
    ex_val := os.ExpandEnv(kv_slice[1])
    env = append(env, kv_slice[0] + "=" + ex_val)
  }
  return env, nil
}
