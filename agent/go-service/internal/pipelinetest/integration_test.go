package pipelinetest

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaaNodeIntegration(t *testing.T) {
	requireIntegration(t)
	root := repositoryRoot(t)
	resource := materializeIntegrationResource(t, root)
	base := Config{
		ImagePath:    filepath.Join(root, "agent", "go-service", "internal", "pipelinetest", "testdata", "positive.png"),
		ResourcePath: resource,
		LibraryDir:   filepath.Join(root, "install", "maafw"),
		AgentMode:    AgentDisabled,
		Expectation:  ExpectHit,
		Output:       OutputJSON,
		AgentTimeout: defaultAgentTimeout,
	}
	tests := []struct {
		name        string
		node        string
		image       string
		expectation Expectation
		algorithm   string
		combined    int
	}{
		{"template and action isolation", "PipelineTestTemplateHit", "positive.png", ExpectHit, "TemplateMatch", 0},
		{"template miss", "PipelineTestTemplateHit", "negative.png", ExpectMiss, "TemplateMatch", 0},
		{"cross-node and", "PipelineTestAnd", "positive.png", ExpectHit, "And", 2},
		{"cross-node or", "PipelineTestOr", "positive.png", ExpectHit, "Or", 2},
		{"ocr hit", "PipelineTestOCR", "positive.png", ExpectHit, "OCR", 0},
		{"ocr miss", "PipelineTestOCR", "negative.png", ExpectMiss, "OCR", 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := base
			config.Node = test.node
			config.ImagePath = filepath.Join(root, "agent", "go-service", "internal", "pipelinetest", "testdata", test.image)
			config.Expectation = test.expectation
			report, err := (runtimeExecutor{stderr: io.Discard}).Execute(config)
			if err != nil {
				t.Fatal(err)
			}
			if report.Outcome != "pass" || report.Algorithm != test.algorithm || len(report.Combined) != test.combined {
				t.Fatalf("unexpected report: %#v", report)
			}
		})
	}

	unknown := base
	unknown.Node = "PipelineTestUnknown"
	_, err := (runtimeExecutor{}).Execute(unknown)
	assertStage(t, err, "node")
}

func TestMaaInvalidResourceIntegration(t *testing.T) {
	requireIntegration(t)
	root := repositoryRoot(t)
	resource := t.TempDir()
	pipelineDir := filepath.Join(resource, "pipeline")
	if err := os.MkdirAll(pipelineDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pipelineDir, "invalid.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := (runtimeExecutor{}).Execute(Config{
		Node:         "PipelineTestTemplateHit",
		ImagePath:    filepath.Join(root, "agent", "go-service", "internal", "pipelinetest", "testdata", "positive.png"),
		ResourcePath: resource,
		LibraryDir:   filepath.Join(root, "install", "maafw"),
		AgentMode:    AgentDisabled,
		Expectation:  ExpectHit,
		Output:       OutputJSON,
		AgentTimeout: defaultAgentTimeout,
	})
	assertStage(t, err, "resource")
}

func TestMaaCustomRecognitionIntegration(t *testing.T) {
	requireIntegration(t)
	root := repositoryRoot(t)
	resource := materializeIntegrationResource(t, root)
	config := Config{
		Node:         "PipelineTestCustom",
		ImagePath:    filepath.Join(root, "agent", "go-service", "internal", "pipelinetest", "testdata", "positive.png"),
		ResourcePath: resource,
		LibraryDir:   filepath.Join(root, "install", "maafw"),
		AgentPath:    filepath.Join(root, "install", "agent", "go-service.exe"),
		AgentMode:    AgentExplicit,
		Expectation:  ExpectHit,
		Output:       OutputJSON,
		AgentTimeout: defaultAgentTimeout,
	}
	report, err := (runtimeExecutor{stderr: io.Discard}).Execute(config)
	if err != nil {
		t.Fatal(err)
	}
	if report.Outcome != "pass" || report.Algorithm != "Custom" {
		t.Fatalf("unexpected report: %#v", report)
	}

	config.AgentMode = AgentDisabled
	_, err = (runtimeExecutor{}).Execute(config)
	assertStage(t, err, "agent")
	if !strings.Contains(err.Error(), "CommonTemplateColorMatchRecognition") {
		t.Fatalf("missing registration diagnostic: %v", err)
	}
}

const defaultAgentTimeout = 10_000_000_000

func requireIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("MDA_PIPELINE_TEST_INTEGRATION") != "1" {
		t.Skip("set MDA_PIPELINE_TEST_INTEGRATION=1 to run MaaFramework integration tests")
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	layout, err := ResolveLayout(cwd)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Dir(filepath.Dir(layout.ResourcePath))
}

func materializeIntegrationResource(t *testing.T, root string) string {
	t.Helper()
	destination := filepath.Join(t.TempDir(), "resource")
	fixture := filepath.Join(root, "agent", "go-service", "internal", "pipelinetest", "testdata", "resource")
	copyTree(t, fixture, destination)
	copyTree(t, filepath.Join(root, "assets", "resource", "model"), filepath.Join(destination, "model"))
	return destination
}

func copyTree(t *testing.T, source, destination string) {
	t.Helper()
	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := os.Link(path, target); err == nil {
			return nil
		}
		input, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, input, info.Mode())
	})
	if err != nil {
		t.Fatal(err)
	}
}
