package pipelinetest

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveLayoutRepository(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "assets", "resource"))
	mustWrite(t, filepath.Join(root, "agent", "go-service", "go.mod"), "module test")

	layout, err := ResolveLayout(filepath.Join(root, "agent", "go-service"))
	if err != nil {
		t.Fatal(err)
	}
	if layout.ResourcePath != filepath.Join(root, "assets", "resource") {
		t.Fatalf("resource = %q", layout.ResourcePath)
	}
	if layout.LibraryDir != filepath.Join(root, "install", "maafw") {
		t.Fatalf("library = %q", layout.LibraryDir)
	}
	if layout.AgentPath != filepath.Join(root, "install", "agent", "go-service.exe") {
		t.Fatalf("agent = %q", layout.AgentPath)
	}
}

func TestResolveLayoutInstall(t *testing.T) {
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "resource"))
	mustMkdir(t, filepath.Join(root, "maafw"))

	layout, err := ResolveLayout(filepath.Join(root, "agent"))
	if err != nil {
		t.Fatal(err)
	}
	if layout.ResourcePath != filepath.Join(root, "resource") || layout.LibraryDir != filepath.Join(root, "maafw") {
		t.Fatalf("unexpected layout: %#v", layout)
	}
}

func TestParseNodeConfigValidation(t *testing.T) {
	root := testRepo(t)
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"node required", []string{"--image", "shot.png"}, "--node is required"},
		{"image required", []string{"--node", "Target"}, "--image is required"},
		{"bad expectation", []string{"--node", "Target", "--image", "shot.png", "--expect", "maybe"}, "must be hit or miss"},
		{"bad output", []string{"--node", "Target", "--image", "shot.png", "--output", "xml"}, "must be text or json"},
		{"agent conflict", []string{"--node", "Target", "--image", "shot.png", "--agent", "agent.exe", "--no-agent"}, "cannot be used together"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseNodeConfig(test.args, root, &bytes.Buffer{})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func TestParseNodeConfigOverrides(t *testing.T) {
	root := testRepo(t)
	config, err := ParseNodeConfig([]string{
		"--node", "Target", "--image", "shot.png", "--resource", "res",
		"--library-dir", "lib", "--agent", "bin\\agent.exe", "--expect", "miss",
		"--output", "json", "--artifacts", "out", "--agent-timeout", "3s",
	}, root, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if config.AgentMode != AgentExplicit || config.Expectation != ExpectMiss || config.Output != OutputJSON {
		t.Fatalf("unexpected config: %#v", config)
	}
	for _, path := range []string{config.ImagePath, config.ResourcePath, config.LibraryDir, config.AgentPath, config.ArtifactDir} {
		if !filepath.IsAbs(path) {
			t.Fatalf("path is not absolute: %q", path)
		}
	}
}

func testRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mustMkdir(t, filepath.Join(root, "assets", "resource"))
	mustWrite(t, filepath.Join(root, "agent", "go-service", "go.mod"), "module test")
	return root
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, path, value string) {
	t.Helper()
	mustMkdir(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
		t.Fatal(err)
	}
}
