package main

import (
	"os"
	"simple-server-template/app/bootstrap"
)

func main() {
	os.Exit(bootstrap.RunCLI())
}
