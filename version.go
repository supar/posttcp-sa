package main

import (
	"fmt"
	"os"
)

func showVersion() {
	fmt.Printf("Postfix tcp map service (%s) %s, built %s\n", NAME, VERSION, BUILDDATE)
	os.Exit(0)
}
