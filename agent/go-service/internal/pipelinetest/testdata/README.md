# Pipeline Test Fixtures

These fixtures are synthetic and owned by `agent/go-service/internal/pipelinetest`.
They contain no user screenshots or game account data.

- `positive.png` is a 1280 by 720 image containing a deterministic marker and the text `ARENA`.
- `negative.png` has the same dimensions without the marker or text.
- `resource/image/*.png` contains marker templates.
- `resource/pipeline/pipeline-test.json` defines built-in and Agent-backed test nodes.

Regenerate image fixtures with `go run ./internal/pipelinetest/testdata/generate` from
`agent/go-service`, then review the resulting binary diff. Update the pipeline and generator
together so fixture intent remains explicit and deterministic.
