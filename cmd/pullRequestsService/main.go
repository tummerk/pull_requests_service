package main

import (
	"fmt"
	"os"
	"pull_requests_service/internal/application"

	_ "go.uber.org/automaxprocs"
)

var appVersion = "v0.0.0"

func main() {
	if err := application.New(appVersion).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
