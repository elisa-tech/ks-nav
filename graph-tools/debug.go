package main

import (
	"fmt"
	"runtime"
	"os"
)

const (
	debugNone            = 0
	debugIO              = 1
	debugAddFunctionName = 15
)

//const DebugLevel uint32 = debugIO | (1<<debugAddFunctionName - 1)
var DebugLevel uint32 = 0

func debugIOPrintf(format string, a ...interface{}) (int, error) {
	var s string
	var n int
	var err error

	if DebugLevel&(1<<(debugIO-1)) != 0 {
		if DebugLevel&(1<<(debugAddFunctionName-1)) != 0 {
			pc, _, _, ok := runtime.Caller(1)
			s = "?"
			if ok {
				fn := runtime.FuncForPC(pc)
				if fn != nil {
					s = fn.Name()
				}
			}
			newformat := "[" + s + "] " + format
			n, err = fmt.Fprintf(os.Stderr, newformat, a...)
		} else {
			n, err = fmt.Fprintf(os.Stderr, format, a...)
		}
		return n, err
	}
	return 0, nil
}
