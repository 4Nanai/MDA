package pipelinetest

import (
	"io"
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentMissingExecutableModes(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing-agent.exe")
	diagnostics := []string{}
	session, err := startAgent(Config{AgentMode: AgentAuto, AgentPath: missing}, nil, &diagnostics, io.Discard)
	if err != nil || session != nil || len(diagnostics) != 1 {
		t.Fatalf("auto mode: session=%v diagnostics=%v error=%v", session, diagnostics, err)
	}
	diagnostics = nil
	_, err = startAgent(Config{AgentMode: AgentExplicit, AgentPath: missing}, nil, &diagnostics, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("explicit mode error = %v", err)
	}
}
