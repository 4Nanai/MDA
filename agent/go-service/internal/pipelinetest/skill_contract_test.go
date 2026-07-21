package pipelinetest

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestPipelineTestingSkillMatchesCLIContract(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	layout, err := ResolveLayout(cwd)
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Dir(filepath.Dir(layout.ResourcePath))
	content, err := os.ReadFile(filepath.Join(root, ".codex", "skills", "pipeline-testing", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	for _, required := range []string{
		"pipeline-test.exe", "--node", "--image", "--resource", "--library-dir",
		"--agent", "--no-agent", "--agent-timeout", "--expect", "--output", "--artifacts",
		"Exit `0`", "Exit `1`", "Exit `2`", "1280 by 720", "pipeline-test/v1",
	} {
		if !strings.Contains(text, required) {
			t.Errorf("skill is missing CLI contract text %q", required)
		}
	}

	allowed := map[string]bool{
		"--node": true, "--image": true, "--resource": true, "--library-dir": true,
		"--agent": true, "--no-agent": true, "--agent-timeout": true,
		"--expect": true, "--output": true, "--artifacts": true,
	}
	for _, flag := range regexp.MustCompile(`--[a-z][a-z-]*`).FindAllString(text, -1) {
		if !allowed[flag] {
			t.Errorf("skill documents unsupported flag %s", flag)
		}
	}
}
