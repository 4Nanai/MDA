package pipelinetest

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

type fakeExecutor struct {
	report Report
	err    error
	config Config
}

func (f *fakeExecutor) Execute(config Config) (Report, error) {
	f.config = config
	return f.report, f.err
}

func TestCLIHelpAndUnknownCommand(t *testing.T) {
	root := testRepo(t)
	tests := []struct {
		args []string
		code int
		want string
	}{
		{[]string{"--help"}, ExitPass, "Commands:"},
		{[]string{"node", "--help"}, ExitPass, "--expect hit|miss"},
		{[]string{"pipeline"}, ExitError, "unknown command"},
	}
	for _, test := range tests {
		var stdout, stderr bytes.Buffer
		code := runCLI(test.args, root, &stdout, &stderr, &fakeExecutor{})
		if code != test.code || !strings.Contains(stdout.String()+stderr.String(), test.want) {
			t.Fatalf("args=%v code=%d output=%q", test.args, code, stdout.String()+stderr.String())
		}
	}
}

func TestCLIExitStatuses(t *testing.T) {
	root := testRepo(t)
	base := []string{"node", "--node", "Target", "--image", "shot.png", "--no-agent"}
	tests := []struct {
		name   string
		report Report
		err    error
		code   int
	}{
		{"pass", Report{Outcome: "pass"}, nil, ExitPass},
		{"assertion", Report{Outcome: "fail"}, nil, ExitAssertion},
		{"runtime", Report{}, &StageError{Stage: "resource", Err: errors.New("bad bundle")}, ExitError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := runCLI(base, root, &stdout, &stderr, &fakeExecutor{report: test.report, err: test.err})
			if code != test.code {
				t.Fatalf("code = %d, want %d", code, test.code)
			}
		})
	}
}

func TestJSONOutputIsOneVersionedDocument(t *testing.T) {
	root := testRepo(t)
	report := Report{SchemaVersion: SchemaVersion, Node: "Target", Outcome: "pass", Detail: json.RawMessage(`{"score":0.9}`)}
	var stdout, stderr bytes.Buffer
	code := runCLI([]string{"node", "--node", "Target", "--image", "shot.png", "--output", "json", "--no-agent"}, root, &stdout, &stderr, &fakeExecutor{report: report})
	if code != ExitPass || stderr.Len() != 0 {
		t.Fatalf("code=%d stderr=%q", code, stderr.String())
	}
	var decoded Report
	if err := json.Unmarshal(stdout.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid JSON: %v; output=%q", err, stdout.String())
	}
	if decoded.SchemaVersion != SchemaVersion || decoded.Node != "Target" {
		t.Fatalf("unexpected report: %#v", decoded)
	}
}

func TestTextOutputIncludesStableFields(t *testing.T) {
	report := Report{Node: "Target", Image: `C:\shot.png`, Expectation: ExpectHit, ActualHit: true, Algorithm: "TemplateMatch", Box: [4]int{1, 2, 3, 4}, DurationMS: 8, Outcome: "pass", Artifacts: []string{`C:\draw.png`}}
	var output bytes.Buffer
	if err := writeReport(OutputText, report, &output); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"node: Target", "actual_hit: true", "box: [1,2,3,4]", "outcome: pass", `artifact: C:\draw.png`} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("missing %q in %q", want, output.String())
		}
	}
}
