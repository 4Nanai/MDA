package pipelinetest

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
)

type runtimeExecutor struct {
	stderr io.Writer
}

type harnessResponse struct {
	detail *maa.RecognitionDetail
	err    error
}

type recognitionHarness struct {
	node   string
	image  image.Image
	result chan harnessResponse
}

func (h *recognitionHarness) Run(ctx *maa.Context, _ *maa.CustomActionArg) bool {
	detail, err := ctx.RunRecognition(h.node, h.image)
	h.result <- harnessResponse{detail: detail, err: err}
	return err == nil
}

func (executor runtimeExecutor) Execute(config Config) (report Report, retErr error) {
	img, err := loadScreenshot(config.ImagePath)
	if err != nil {
		return Report{}, &StageError{Stage: "image", Err: err}
	}
	if !isDir(config.LibraryDir) {
		return Report{}, &StageError{Stage: "framework", Err: fmt.Errorf("library directory does not exist: %s", config.LibraryDir)}
	}
	if !isDir(config.ResourcePath) {
		return Report{}, &StageError{Stage: "resource", Err: fmt.Errorf("bundle directory does not exist: %s", config.ResourcePath)}
	}
	if config.ArtifactDir != "" {
		if err := os.MkdirAll(config.ArtifactDir, 0o755); err != nil {
			return Report{}, &StageError{Stage: "artifacts", Err: err}
		}
	}

	options := []maa.InitOption{
		maa.WithLibDir(config.LibraryDir),
		maa.WithJSONEncoder(json.Marshal),
		maa.WithJSONDecoder(json.Unmarshal),
	}
	if config.ArtifactDir != "" {
		options = append(options,
			maa.WithDebugMode(true),
			maa.WithSaveDraw(true),
			maa.WithLogDir(config.ArtifactDir),
		)
	}
	if err := maa.Init(options...); err != nil {
		return Report{}, &StageError{Stage: "framework", Err: fmt.Errorf("initialize from %s: %w", config.LibraryDir, err)}
	}
	defer func() {
		if err := maa.Release(); err != nil && retErr == nil {
			retErr = &StageError{Stage: "framework", Err: fmt.Errorf("release: %w", err)}
		}
	}()
	configDir := filepath.Join(os.TempDir(), "mda-pipeline-test")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return Report{}, &StageError{Stage: "framework", Err: fmt.Errorf("create toolkit config directory: %w", err)}
	}
	if err := maa.ConfigInitOption(configDir, "{}"); err != nil {
		return Report{}, &StageError{Stage: "framework", Err: fmt.Errorf("initialize toolkit options: %w", err)}
	}

	resource, err := maa.NewResource()
	if err != nil {
		return Report{}, &StageError{Stage: "resource", Err: err}
	}
	defer resource.Destroy()
	if !resource.PostBundle(config.ResourcePath).Wait().Success() {
		return Report{}, &StageError{Stage: "resource", Err: fmt.Errorf("failed to load bundle: %s", config.ResourcePath)}
	}
	if _, err := resource.GetNodeJSON(config.Node); err != nil {
		return Report{}, &StageError{Stage: "node", Err: fmt.Errorf("unknown node %q: %w", config.Node, err)}
	}

	diagnostics := make([]string, 0, 2)
	agentOutput := executor.stderr
	if agentOutput == nil {
		agentOutput = io.Discard
	}
	agent, err := startAgent(config, resource, &diagnostics, agentOutput)
	if err != nil {
		return Report{}, &StageError{Stage: "agent", Err: err}
	}
	if agent != nil {
		defer agent.close()
	}
	required, err := requiredCustomRecognitions(resource, config.Node)
	if err != nil {
		return Report{}, &StageError{Stage: "node", Err: fmt.Errorf("inspect Custom Recognition dependencies: %w", err)}
	}
	missing, err := missingRegistrations(resource, required)
	if err != nil {
		return Report{}, &StageError{Stage: "agent", Err: fmt.Errorf("read Resource registrations: %w", err)}
	}
	if len(missing) != 0 {
		return Report{}, &StageError{Stage: "agent", Err: fmt.Errorf("required Custom Recognition registration(s) unavailable: %s", strings.Join(missing, ", "))}
	}

	controller, err := maa.NewBlankController()
	if err != nil {
		return Report{}, &StageError{Stage: "controller", Err: err}
	}
	defer controller.Destroy()
	if !controller.PostConnect().Wait().Success() {
		return Report{}, &StageError{Stage: "controller", Err: fmt.Errorf("BlankController connection failed")}
	}
	tasker, err := maa.NewTasker()
	if err != nil {
		return Report{}, &StageError{Stage: "tasker", Err: err}
	}
	defer tasker.Destroy()
	if err := tasker.BindResource(resource); err != nil {
		return Report{}, &StageError{Stage: "tasker", Err: err}
	}
	if err := tasker.BindController(controller); err != nil {
		return Report{}, &StageError{Stage: "tasker", Err: err}
	}
	if !tasker.Initialized() {
		return Report{}, &StageError{Stage: "tasker", Err: fmt.Errorf("Tasker is not initialized")}
	}

	harnessName := fmt.Sprintf("__PipelineTestHarness_%d_%d", os.Getpid(), time.Now().UnixNano())
	entryName := harnessName + "_Entry"
	harness := &recognitionHarness{node: config.Node, image: img, result: make(chan harnessResponse, 1)}
	if err := resource.RegisterCustomAction(harnessName, harness); err != nil {
		return Report{}, &StageError{Stage: "harness", Err: err}
	}

	zero := time.Duration(0)
	pipeline := maa.NewPipeline().AddNode(
		maa.NewNode(entryName).
			SetRecognition(maa.RecDirectHit()).
			SetAction(maa.ActCustom(maa.CustomActionParam{CustomAction: harnessName})).
			SetRateLimit(zero).
			SetPreDelay(zero).
			SetPostDelay(zero),
	)
	started := time.Now()
	job := tasker.PostTask(entryName, pipeline).Wait()
	if err := job.Error(); err != nil {
		return Report{}, &StageError{Stage: "harness", Err: err}
	}
	select {
	case response := <-harness.result:
		if response.err != nil {
			return Report{}, &StageError{Stage: "recognition", Err: response.err}
		}
		if response.detail == nil {
			return Report{}, &StageError{Stage: "recognition", Err: fmt.Errorf("MaaFramework returned no detail")}
		}
		report = reportFromDetail(config, response.detail, time.Since(started))
		report.Diagnostics = diagnostics
		if config.ArtifactDir != "" {
			paths, saveErr := saveDrawArtifacts(config.ArtifactDir, response.detail)
			if saveErr != nil {
				return Report{}, &StageError{Stage: "artifacts", Err: saveErr}
			}
			report.Artifacts = paths
		}
		return report, nil
	default:
		if !job.Success() {
			return Report{}, &StageError{Stage: "harness", Err: fmt.Errorf("private recognition task failed")}
		}
		return Report{}, &StageError{Stage: "harness", Err: fmt.Errorf("private recognition task completed without a result")}
	}
}

func saveDrawArtifacts(dir string, detail *maa.RecognitionDetail) ([]string, error) {
	var paths []string
	var visit func(*maa.RecognitionDetail) error
	index := 0
	visit = func(current *maa.RecognitionDetail) error {
		if current == nil {
			return nil
		}
		for drawIndex, draw := range current.Draws {
			name := fmt.Sprintf("%02d-%s-%02d.png", index, sanitizeFilename(current.Name), drawIndex)
			path := filepath.Join(dir, name)
			file, err := os.Create(path)
			if err != nil {
				return err
			}
			encodeErr := png.Encode(file, draw)
			closeErr := file.Close()
			if encodeErr != nil {
				return encodeErr
			}
			if closeErr != nil {
				return closeErr
			}
			paths = append(paths, path)
			index++
		}
		for _, child := range current.CombinedResult {
			if err := visit(child); err != nil {
				return err
			}
		}
		return nil
	}
	return paths, visit(detail)
}

func sanitizeFilename(value string) string {
	if value == "" {
		return "recognition"
	}
	return strings.Map(func(r rune) rune {
		switch r {
		case '<', '>', ':', '"', '/', '\\', '|', '?', '*':
			return '_'
		default:
			return r
		}
	}, value)
}
