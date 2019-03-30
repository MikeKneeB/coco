package backend

import (
  "strings"
  "strconv"
  "regexp"
  "fmt"
  //"path/filepath"
)

// Golang specific regexp stuff - this needs to go

var fullReg string = `(?m)(%s.+?):(\d+?):\d+?:(.*?)`

type CompileLine struct {
  FileName string
  Line int
  Message string
}

func Parse(output string, root string) []CompileLine {
  re, err := regexp.Compile(fmt.Sprintf(fullReg, root))
  cls := make([]CompileLine, 0, strings.Count(output, "\n"))
  if err != nil {
    return cls
  }

  matches := re.FindAllStringSubmatchIndex(output, -1)

  for i, match := range matches {
    cls = append(cls, CompileLine{})
    cls[i].FileName = output[match[2]:match[3]]
    cls[i].Line, _ = strconv.Atoi(output[match[4]:match[5]])
    if i + 1 != len(matches) {
      cls[i].Message = output[match[6]:matches[i+1][0]]
    } else {
      cls[i].Message = output[match[6]:]
    }
  }

  return cls
}
