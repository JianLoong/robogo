package main

import (
	"fmt"
	"os"

	"github.com/JianLoong/robogo/internal/cli"
)

var (
	version = "0.1.0"
	commit  = "dev"
	date    = "unknown"
)

func main() {
	if err := realMain(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func realMain() error {
	app := cli.NewApp(version, commit, date)
	return app.Run(os.Args)
}