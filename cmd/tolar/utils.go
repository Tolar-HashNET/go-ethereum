package main

import (
	"fmt"
	"os"
)

func HandleFatalError(msg string, error error) {
	if error != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v: %v\n", msg, error)
		os.Exit(1)
	}
}

func ThrowFatalError(msg string) {
	_, _ = fmt.Fprintf(os.Stderr, "%v\n", msg)
	os.Exit(1)
}
