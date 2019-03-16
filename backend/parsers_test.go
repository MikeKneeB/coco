package backend

import (
  "testing"
  //"regexp"
)

var pretend_out string = `
# github.com/MikeKneeB/coco/backend
src/github.com/MikeKneeB/coco/backend/parsers.go:28:21: undefined: output_split
src/github.com/MikeKneeB/coco/backend/parsers.go:30:17: undefined: fullNameReg
src/github.com/MikeKneeB/coco/backend/parsers.go:33:7: cl.FilePath undefined (type CompileLine has no field or method FilePath)
src/github.com/MikeKneeB/coco/backend/parsers.go:34:16: cannot assign int64 to cl.Line (type int) in multiple assignment
and on a new line
src/github.com/MikeKneeB/coco/backend/parsers.go:34:35: undefined: lineReg
src/github.com/MikeKneeB/coco/backend/parsers.go:35:18: undefined: msgReg
src/github.com/MikeKneeB/coco/backend/parsers.go:36:7: cl.MessageLines undefined (type CompileLine has no field or method MessageLines)
`

func TestRegExps(t *testing.T) {

  matches := fullReg.FindAllStringSubmatchIndex(pretend_out, -1)

  t.Log(len(matches))
  t.Log(matches)

  for i, match := range matches {
    t.Log(pretend_out[match[2]:match[3]])
    t.Log(pretend_out[match[4]:match[5]])
    if i + 1 != len(matches) {
      t.Log(pretend_out[match[6]:matches[i+1][0]])
    } else {
      t.Log(pretend_out[match[6]:])
    }
  }

}
