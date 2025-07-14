package main

import (
	"github.com/JianLoong/robogo/internal"
)

var (
	version = "1.0.0-simplified"
	commit  = "simplified"
	date    = "2025"
)

func main() {
	// Run the simplified CLI - no abstractions, just direct execution
	internal.RunCLI()
}