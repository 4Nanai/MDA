package main

import (
	"os"

	"github.com/1204244136/MDA/agent/go-service/internal/pipelinetest"
)

func main() {
	os.Exit(pipelinetest.RunCLI(os.Args[1:], os.Stdout, os.Stderr))
}
