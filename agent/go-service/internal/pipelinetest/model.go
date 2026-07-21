package pipelinetest

import (
	"encoding/json"
	"fmt"
	"time"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
)

const SchemaVersion = "pipeline-test/v1"

type RecognitionResult struct {
	Node      string              `json:"node"`
	Algorithm string              `json:"algorithm"`
	Hit       bool                `json:"hit"`
	Box       [4]int              `json:"box"`
	Detail    json.RawMessage     `json:"detail"`
	Combined  []RecognitionResult `json:"combined,omitempty"`
}

type Report struct {
	SchemaVersion string              `json:"schema_version"`
	Node          string              `json:"node"`
	Image         string              `json:"image"`
	Expectation   Expectation         `json:"expectation"`
	ActualHit     bool                `json:"actual_hit"`
	Algorithm     string              `json:"algorithm"`
	Box           [4]int              `json:"box"`
	DurationMS    int64               `json:"duration_ms"`
	Outcome       string              `json:"outcome"`
	Detail        json.RawMessage     `json:"detail"`
	Combined      []RecognitionResult `json:"combined,omitempty"`
	Artifacts     []string            `json:"artifacts,omitempty"`
	Diagnostics   []string            `json:"diagnostics,omitempty"`
}

type StageError struct {
	Stage string `json:"stage"`
	Err   error  `json:"-"`
}

func (e *StageError) Error() string { return fmt.Sprintf("%s: %v", e.Stage, e.Err) }
func (e *StageError) Unwrap() error { return e.Err }

type ErrorReport struct {
	SchemaVersion string `json:"schema_version"`
	Outcome       string `json:"outcome"`
	Stage         string `json:"stage"`
	Error         string `json:"error"`
}

func reportFromDetail(config Config, detail *maa.RecognitionDetail, elapsed time.Duration) Report {
	result := convertDetail(detail)
	outcome := "fail"
	if (config.Expectation == ExpectHit) == detail.Hit {
		outcome = "pass"
	}
	return Report{
		SchemaVersion: SchemaVersion,
		Node:          config.Node,
		Image:         config.ImagePath,
		Expectation:   config.Expectation,
		ActualHit:     detail.Hit,
		Algorithm:     detail.Algorithm,
		Box:           [4]int(detail.Box),
		DurationMS:    elapsed.Milliseconds(),
		Outcome:       outcome,
		Detail:        result.Detail,
		Combined:      result.Combined,
	}
}

func convertDetail(detail *maa.RecognitionDetail) RecognitionResult {
	if detail == nil {
		return RecognitionResult{Detail: json.RawMessage("null")}
	}
	result := RecognitionResult{
		Node:      detail.Name,
		Algorithm: detail.Algorithm,
		Hit:       effectiveHit(detail),
		Box:       [4]int(detail.Box),
		Detail:    normalizeJSON(detail.DetailJson),
	}
	for _, child := range detail.CombinedResult {
		result.Combined = append(result.Combined, convertDetail(child))
	}
	return result
}

func effectiveHit(detail *maa.RecognitionDetail) bool {
	if detail == nil {
		return false
	}
	if detail.Hit {
		return true
	}
	if detail.Results != nil {
		return detail.Results.Best != nil
	}
	switch detail.Algorithm {
	case string(maa.RecognitionTypeDirectHit):
		return true
	case string(maa.RecognitionTypeAnd):
		if len(detail.CombinedResult) == 0 {
			return false
		}
		for _, child := range detail.CombinedResult {
			if !effectiveHit(child) {
				return false
			}
		}
		return true
	case string(maa.RecognitionTypeOr):
		for _, child := range detail.CombinedResult {
			if effectiveHit(child) {
				return true
			}
		}
	}
	return false
}

func normalizeJSON(value string) json.RawMessage {
	raw := json.RawMessage(value)
	if len(raw) == 0 || !json.Valid(raw) {
		return json.RawMessage("null")
	}
	return raw
}
