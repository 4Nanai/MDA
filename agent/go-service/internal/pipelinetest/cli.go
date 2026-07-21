package pipelinetest

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

const (
	ExitPass      = 0
	ExitAssertion = 1
	ExitError     = 2
)

type Executor interface {
	Execute(Config) (Report, error)
}

func RunCLI(args []string, stdout, stderr io.Writer) int {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "pipeline-test: %v\n", err)
		return ExitError
	}
	return runCLI(args, cwd, stdout, stderr, runtimeExecutor{stderr: stderr})
}

func runCLI(args []string, cwd string, stdout, stderr io.Writer, executor Executor) int {
	if len(args) == 0 {
		writeRootHelp(stderr)
		return ExitError
	}
	if args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		writeRootHelp(stdout)
		return ExitPass
	}
	if args[0] != "node" {
		fmt.Fprintf(stderr, "pipeline-test: unknown command %q\n\n", args[0])
		writeRootHelp(stderr)
		return ExitError
	}
	if len(args) == 2 && (args[1] == "-h" || args[1] == "--help") {
		writeNodeHelp(stdout, nil)
		return ExitPass
	}

	config, err := ParseNodeConfig(args[1:], cwd, stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitPass
		}
		fmt.Fprintf(stderr, "pipeline-test node: %v\n\n", err)
		writeNodeHelp(stderr, nil)
		return ExitError
	}
	report, err := executor.Execute(config)
	if err != nil {
		writeExecutionError(config.Output, err, stdout, stderr)
		return ExitError
	}
	if err := writeReport(config.Output, report, stdout); err != nil {
		fmt.Fprintf(stderr, "pipeline-test output: %v\n", err)
		return ExitError
	}
	if report.Outcome == "pass" {
		return ExitPass
	}
	return ExitAssertion
}

func writeRootHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: pipeline-test <command> [options]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  node    Test one pipeline node's recognition against a local screenshot")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Run 'pipeline-test node --help' for node options.")
}

func writeNodeHelp(w io.Writer, fs *flag.FlagSet) {
	fmt.Fprintln(w, "Usage: pipeline-test node --node NAME --image PATH [options]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Tests recognition only; the target action and pipeline flow are never run.")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Options:")
	if fs != nil {
		fs.PrintDefaults()
		return
	}
	fmt.Fprintln(w, "  --node NAME             pipeline node name (required)")
	fmt.Fprintln(w, "  --image PATH            local 1280x720 PNG or JPEG (required)")
	fmt.Fprintln(w, "  --resource PATH         resource bundle directory")
	fmt.Fprintln(w, "  --library-dir PATH      MaaFramework library directory")
	fmt.Fprintln(w, "  --agent PATH            explicit Agent executable")
	fmt.Fprintln(w, "  --no-agent              disable Agent startup")
	fmt.Fprintln(w, "  --agent-timeout VALUE   startup/connect timeout (default 10s)")
	fmt.Fprintln(w, "  --expect hit|miss       expected result (default hit)")
	fmt.Fprintln(w, "  --output text|json      output format (default text)")
	fmt.Fprintln(w, "  --artifacts PATH        save available draw images")
}

func writeReport(format OutputFormat, report Report, w io.Writer) error {
	if format == OutputJSON {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}
	fmt.Fprintf(w, "node: %s\n", report.Node)
	fmt.Fprintf(w, "image: %s\n", report.Image)
	fmt.Fprintf(w, "expectation: %s\n", report.Expectation)
	fmt.Fprintf(w, "actual_hit: %t\n", report.ActualHit)
	fmt.Fprintf(w, "algorithm: %s\n", report.Algorithm)
	fmt.Fprintf(w, "box: [%d,%d,%d,%d]\n", report.Box[0], report.Box[1], report.Box[2], report.Box[3])
	fmt.Fprintf(w, "duration_ms: %d\n", report.DurationMS)
	fmt.Fprintf(w, "outcome: %s\n", report.Outcome)
	for _, diagnostic := range report.Diagnostics {
		fmt.Fprintf(w, "diagnostic: %s\n", diagnostic)
	}
	for _, artifact := range report.Artifacts {
		fmt.Fprintf(w, "artifact: %s\n", artifact)
	}
	return nil
}

func writeExecutionError(format OutputFormat, err error, stdout, stderr io.Writer) {
	stage := "execution"
	var staged *StageError
	if errors.As(err, &staged) {
		stage = staged.Stage
	}
	if format == OutputJSON {
		_ = json.NewEncoder(stdout).Encode(ErrorReport{
			SchemaVersion: SchemaVersion,
			Outcome:       "error",
			Stage:         stage,
			Error:         err.Error(),
		})
		return
	}
	fmt.Fprintf(stderr, "pipeline-test %s\n", err)
}
