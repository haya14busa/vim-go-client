package vim

import (
	"log"
	"os"
)

var logger *log.Logger = log.New(os.Stderr, "", log.LstdFlags)

// SetLogger sets the logger that is used in vim server. Call only from
// init() functions.
func SetLogger(l *log.Logger) {
	logger = l
}
