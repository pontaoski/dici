package cmd

import (
	"fmt"
	"os"
)

// Error displays a formatted string and then exists
func Error(out string, v ...interface{}) {
	fmt.Printf(out+"\n", v...)
	os.Exit(1)
}
