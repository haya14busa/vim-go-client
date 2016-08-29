package vim

import (
	"log"
	"os"
)

var logger = log.New(os.Stderr, "", log.LstdFlags)

// SetLogger sets the logger that is used in go process. Call only from
// init() functions.
func SetLogger(l *log.Logger) {
	logger = l
}
