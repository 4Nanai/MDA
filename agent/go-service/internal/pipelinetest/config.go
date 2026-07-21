package pipelinetest

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Expectation string

const (
	ExpectHit  Expectation = "hit"
	ExpectMiss Expectation = "miss"
)

type OutputFormat string

const (
	OutputText OutputFormat = "text"
	OutputJSON OutputFormat = "json"
)

type AgentMode string

const (
	AgentAuto     AgentMode = "auto"
	AgentExplicit AgentMode = "explicit"
	AgentDisabled AgentMode = "disabled"
)

type Config struct {
	Node         string
	ImagePath    string
	ResourcePath string
	LibraryDir   string
	AgentPath    string
	AgentMode    AgentMode
	Expectation  Expectation
	Output       OutputFormat
	ArtifactDir  string
	AgentTimeout time.Duration
}

type Layout struct {
	ResourcePath string
	LibraryDir   string
	AgentPath    string
}

func ResolveLayout(start string) (Layout, error) {
	abs, err := filepath.Abs(start)
	if err != nil {
		return Layout{}, err
	}
	if info, statErr := os.Stat(abs); statErr == nil && !info.IsDir() {
		abs = filepath.Dir(abs)
	}

	for dir := abs; ; dir = filepath.Dir(dir) {
		repoResource := filepath.Join(dir, "assets", "resource")
		if isDir(repoResource) && fileExists(filepath.Join(dir, "agent", "go-service", "go.mod")) {
			return Layout{
				ResourcePath: repoResource,
				LibraryDir:   filepath.Join(dir, "install", "maafw"),
				AgentPath:    filepath.Join(dir, "install", "agent", "go-service.exe"),
			}, nil
		}

		installResource := filepath.Join(dir, "resource")
		if isDir(installResource) && isDir(filepath.Join(dir, "maafw")) {
			return Layout{
				ResourcePath: installResource,
				LibraryDir:   filepath.Join(dir, "maafw"),
				AgentPath:    filepath.Join(dir, "agent", "go-service.exe"),
			}, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return Layout{}, fmt.Errorf("could not find repository root or install layout from %q", start)
}

func ParseNodeConfig(args []string, startDir string, output io.Writer) (Config, error) {
	layout, err := ResolveLayout(startDir)
	if err != nil {
		return Config{}, err
	}

	fs := flag.NewFlagSet("pipeline-test node", flag.ContinueOnError)
	fs.SetOutput(output)
	fs.Usage = func() { writeNodeHelp(output, fs) }

	config := Config{Expectation: ExpectHit, Output: OutputText, AgentTimeout: 10 * time.Second}
	var noAgent bool
	fs.StringVar(&config.Node, "node", "", "pipeline node name (required)")
	fs.StringVar(&config.ImagePath, "image", "", "local PNG or JPEG screenshot (required)")
	fs.StringVar(&config.ResourcePath, "resource", layout.ResourcePath, "MaaFramework resource bundle directory")
	fs.StringVar(&config.LibraryDir, "library-dir", layout.LibraryDir, "MaaFramework library directory")
	fs.StringVar(&config.AgentPath, "agent", "", "Agent executable path (default: auto-detect)")
	fs.BoolVar(&noAgent, "no-agent", false, "disable Agent startup")
	fs.Var((*expectationValue)(&config.Expectation), "expect", "recognition expectation: hit or miss")
	fs.Var((*outputValue)(&config.Output), "output", "output format: text or json")
	fs.StringVar(&config.ArtifactDir, "artifacts", "", "directory for MaaFramework draw artifacts")
	fs.DurationVar(&config.AgentTimeout, "agent-timeout", config.AgentTimeout, "Agent startup/connect timeout")
	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}
	if fs.NArg() != 0 {
		return Config{}, fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}
	if config.Node == "" {
		return Config{}, errors.New("--node is required")
	}
	if config.ImagePath == "" {
		return Config{}, errors.New("--image is required")
	}
	if config.AgentTimeout <= 0 {
		return Config{}, errors.New("--agent-timeout must be greater than zero")
	}
	if noAgent && config.AgentPath != "" {
		return Config{}, errors.New("--agent and --no-agent cannot be used together")
	}

	config.ImagePath = cleanAbs(config.ImagePath, startDir)
	config.ResourcePath = cleanAbs(config.ResourcePath, startDir)
	config.LibraryDir = cleanAbs(config.LibraryDir, startDir)
	if config.ArtifactDir != "" {
		config.ArtifactDir = cleanAbs(config.ArtifactDir, startDir)
	}
	switch {
	case noAgent:
		config.AgentMode = AgentDisabled
	case config.AgentPath != "":
		config.AgentMode = AgentExplicit
		config.AgentPath = cleanAbs(config.AgentPath, startDir)
	default:
		config.AgentMode = AgentAuto
		config.AgentPath = layout.AgentPath
	}
	return config, nil
}

type expectationValue Expectation

func (v *expectationValue) String() string { return string(*v) }
func (v *expectationValue) Set(value string) error {
	switch Expectation(value) {
	case ExpectHit, ExpectMiss:
		*v = expectationValue(value)
		return nil
	default:
		return fmt.Errorf("must be hit or miss, got %q", value)
	}
}

type outputValue OutputFormat

func (v *outputValue) String() string { return string(*v) }
func (v *outputValue) Set(value string) error {
	switch OutputFormat(value) {
	case OutputText, OutputJSON:
		*v = outputValue(value)
		return nil
	default:
		return fmt.Errorf("must be text or json, got %q", value)
	}
}

func cleanAbs(path, base string) string {
	if !filepath.IsAbs(path) {
		path = filepath.Join(base, path)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return filepath.Clean(abs)
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
