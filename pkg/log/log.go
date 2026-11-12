package log

import (
	"log"
	"os"
)

var std = log.New(os.Stdout, "", log.LstdFlags|log.LUTC)

func Printf(format string, args ...any) {
	std.Printf(format, args...)
}
