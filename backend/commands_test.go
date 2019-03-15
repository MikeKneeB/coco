package backend

import (
  "testing"
  "os"
  "os/exec"
  "strings"
  "io/ioutil"
  "bufio"
)

func TestRoutineBasic(t *testing.T) {
  c := exec.Command("echo", "Hello, world!")
  out_chan := make(chan RoutineOut)
  go RunRoutine(out_chan, c)
  res_string, err := (<-out_chan)()
  t.Log(strings.TrimSpace(res_string))
  t.Log(err)
  if (res_string != "Hello, world!\n") || (err != nil) {
    t.Fail()
  }
}

func TestRoutineNoCommand(t *testing.T) {
  c := exec.Command("echo_but_not", "Hello, world!")
  out_chan := make(chan RoutineOut)
  go RunRoutine(out_chan, c)
  res_string, err := (<-out_chan)()
  t.Log(strings.TrimSpace(res_string))
  t.Log(err)
  if err == nil {
    t.Fail()
  }
}

func TestRoutineCD(t *testing.T) {
  err := os.MkdirAll("/tmp/go_tests", 0755)
  if err != nil {
    t.Log(err)
    t.Fail()
  }
  err = ioutil.WriteFile("/tmp/go_tests/test_file.txt", []byte("A test file\n"),
    0644)
  c := exec.Command("cat", "test_file.txt")
  c.Dir = "/tmp/go_tests"
  out_chan := make(chan RoutineOut)
  go RunRoutine(out_chan, c)
  res_string, err := (<-out_chan)()
  t.Log(strings.TrimSpace(res_string))
  t.Log(err)
  if (res_string != "A test file\n") || (err != nil) {
    t.Fail()
  }
}

func TestContinual(t *testing.T) {
  out_chan := make(chan RoutineOut)
  quit_chan := make(chan bool)
  com_chan := make(chan *exec.Cmd)
  go ContinualRoutine(out_chan, quit_chan, com_chan)
  for i := 0; i != 2; i++ {
    c := exec.Command("echo", "Hello, world!")
    com_chan <- c
    res_string, err := (<-out_chan)()
    t.Log(strings.TrimSpace(res_string))
    t.Log(err)
    if (res_string != "Hello, world!\n") || (err != nil) {
      t.Fail()
    }
  }
  quit_chan <- true
  res_string, err := (<-out_chan)()
  t.Log(strings.TrimSpace(res_string))
  t.Log(err)
  if (res_string != "" || err != ErrRoutineQuit) {
    t.Fail()
  }
}

func TestAddEnv(t *testing.T) {
  envs := []string{"FIRST_ENV=THIS", "SECOND_ENV=THAT", "A_PATH=/tmp/path"}
  temp_env := os.Environ()
  defer resetEnv(temp_env)
  resetEnv(envs)
  to_add := []string{"MORE_ENVS=MORE", "EXPAND=$A_PATH:/tmp/more",
                     "ALSO_EXP=/tmp/some/$SECOND_ENV"}
  expected := []string{"FIRST_ENV=THIS", "SECOND_ENV=THAT", "A_PATH=/tmp/path",
                       "MORE_ENVS=MORE", "EXPAND=/tmp/path:/tmp/more",
                       "ALSO_EXP=/tmp/some/THAT"}
  result, err := MakeEnvironment(to_add)
  if err != nil {
    t.Error(err)
  }
  t.Log(result)
  if !unorderSliceEqual(result, expected) {
    t.Fail()
  }
}

func TestAddEnvBadConfig(t *testing.T) {
  to_add := []string{"NOT VALID"}
  result, err := MakeEnvironment(to_add)
  if result != nil && err == nil {
    t.Fail()
  }
  t.Log(err)
  to_add = []string{"TOO=MANY=EQUALS"}
  result, err = MakeEnvironment(to_add)
  if result != nil && err == nil {
    t.Fail()
  }
  t.Log(err)
}

func TestRunInit(t *testing.T) {
  env := []string{"TEST_ENV=Hello, world!"}
  env, err := MakeEnvironment(env)
  if err != nil {
    t.Error(err)
  }
  args := []string{"argument 1"}
  err = RunInit("./test_util/test_script", ".", args, env)
  if err != nil {
    t.Error(err)
  }
  _, err = os.Stat("out.txt")
  if err != nil {
    t.Error(err)
  }
  defer os.Remove("out.txt")
  f, err := os.Open("out.txt")
  if err != nil {
    t.Error(err)
  }
  defer f.Close()
  scanner := bufio.NewScanner(f)
  scanner.Scan()
  t.Log(scanner.Text())
  if scanner.Text() != "argument 1" {
    t.Fail()
  }
  scanner.Scan()
  t.Log(scanner.Text())
  if scanner.Text() != "Hello, world!" {
    t.Fail()
  }
}

func TestRunInitBadConf(t *testing.T) {
  env := []string{"NOT VALID"}
  args := []string{"argument 1"}
  err := RunInit("./test_script", ".", args, env)
  t.Log(err)
  if err == nil {
    t.Fail();
  }
}

func TestCommandDefString(t *testing.T) {
  c := CommandDef{"hello", "world", []string{"these", "are", "args"}, []string{"this", "is", "env"}}
  if c.String() != "hello these are args" {
    t.Error("Command string", c.String(), "expected: hello these are args")
  }
}

func TestMakeRunnable(t *testing.T) {
  craw := CommandDef{"hello", "world", []string{"these", "are", "args"}, []string{"this", "is", "env"}}
  c := craw.MakeRunnable()

  if c.Path != "hello" {
    t.Error("Cmd path", c.Path, "expected hello")
  }

  if !unorderSliceEqual(c.Args, []string{"hello", "these", "are", "args"}) {
    t.Error("Cmd args", c.Args, "expected", []string{"hello", "these", "are", "args"})
  }

  if c.Dir != "world" {
    t.Error("Cmd dir", c.Dir, "expected world")
  }

  if !unorderSliceEqual(c.Env, []string{"this", "is", "env"}) {
    t.Error("Cmd env", c.Env, "expected", []string{"this", "is", "env"})
  }
}

func unorderSliceEqual(a, b []string) bool {
  if len(a) != len(b) {
    return false
  }
  for _, v := range a {
     c := 0
    for _, ov := range b {
      if v == ov {
        c = c + 1
      }
    }
    if c != 1 {
      return false
    }
  }
  return true
}

func resetEnv(env []string) {
  os.Clearenv()
  for _, v := range env {
    kv_slice := strings.Split(v, "=")
    os.Setenv(kv_slice[0], kv_slice[1])
  }
}
